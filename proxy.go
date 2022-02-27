// Copyright © 2020 Daniel Hodges <hodges.daniel.scott@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/syzkaller/dashboard/dashapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	errUnknownMethod = errors.New("unknown method")
)

// Proxy is a syzkaller dashboard proxy.
type Proxy interface {
	Proxy(*gin.Context)
	Metrics(*gin.Context)
}

type proxy struct {
	dashMu sync.RWMutex
	dashes map[string]*dashapi.Dashboard
}

// New returns a new proxy
func New(forward []string) Proxy {
	dashes := map[string]*dashapi.Dashboard{}
	for _, f := range forward {
		dashes[f] = dashapi.New("proxy", f, "")
	}
	return &proxy{
		dashes: dashes,
	}
}

// Metrics implements the metrics interface.
func (p *proxy) Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// Proxy implements the Proxy interface.
func (p *proxy) Proxy(c *gin.Context) {
	client := c.PostForm("client")
	key := c.PostForm("key")
	method := c.PostForm("method")
	fmt.Printf("client: %q, key: %q, method: %q", client, key, method)

	switch method {
	case "upload_build":
		p.uploadBuild(c, client, key)
	case "builder_poll":
		p.builderPoll(c, client, key)
	case "job_poll":
		p.jobPoll(c, client, key)
	case "job_done":
		p.jobDone(c, client, key)
	case "report_build_error":
		p.reportBuildError(c, client, key)
	case "commit_poll":
		p.commitPoll(c, client, key)
	case "upload_commits":
		p.uploadCommits(c, client, key)
	case "report_crash":
		p.reportCrash(c, client, key)
	case "need_repro":
		p.needRepro(c, client, key)
	case "report_failed_repro":
		p.reportFailedRepro(c, client, key)
	case "log_error":
		p.logError(c, client, key)
	case "reporting_poll_bugs":
		p.reportingPollBugs(c, client, key)
	case "reporting_poll_notifs":
		p.reportingPollNotifs(c, client, key)
	case "reporting_poll_closed":
		p.reportingPollClosed(c, client, key)
	case "reporting_update":
		p.reportingUpdate(c, client, key)
	case "manager_stats":
		p.managerStats(c, client, key)
		rpcCounters.WithLabelValues(client, method).Inc()
		return
	case "bug_list":
		p.bugList(c, client, key)
	case "load_bug":
		p.loadBug(c, client, key)
	default:
		rpcCounters.WithLabelValues(client, "invalid").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	rpcCounters.WithLabelValues(client, method).Inc()
}

func (p *proxy) uploadBuild(c *gin.Context, client, key string) {
	var (
		build dashapi.Build
		// Payload is gzip'd json.
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&build); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.UploadBuild(&build)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) builderPoll(c *gin.Context, client, key string) {
	var (
		pollReq dashapi.BuilderPollReq
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&pollReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.BuilderPoll(pollReq.Manager)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) jobPoll(c *gin.Context, client, key string) {
	var (
		jobPollReq dashapi.JobPollReq
		payload    = c.PostForm("payload")
		buf        = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&jobPollReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.JobPoll(&jobPollReq)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) jobDone(c *gin.Context, client, key string) {
	var (
		jobDoneReq dashapi.JobDoneReq
		payload    = c.PostForm("payload")
		buf        = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&jobDoneReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.JobDone(&jobDoneReq)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportBuildError(c *gin.Context, client, key string) {
	var (
		buildErrReq dashapi.BuildErrorReq
		payload     = c.PostForm("payload")
		buf         = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&buildErrReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.ReportBuildError(&buildErrReq)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) commitPoll(c *gin.Context, client, key string) {
	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.CommitPoll()
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) uploadCommits(c *gin.Context, client, key string) {
	var (
		req     dashapi.CommitPollResultReq
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.UploadCommits(req.Commits)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportCrash(c *gin.Context, client, key string) {
	var (
		req     dashapi.Crash
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.ReportCrash(&req)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) needRepro(c *gin.Context, client, key string) {
	var (
		req     dashapi.CrashID
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.NeedRepro(&req)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportFailedRepro(c *gin.Context, client, key string) {
	var (
		req     dashapi.CrashID
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.ReportFailedRepro(&req)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) logError(c *gin.Context, client, key string) {
	var (
		req     dashapi.LogEntry
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		dash.LogError(req.Name, req.Text)
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportingPollBugs(c *gin.Context, client, key string) {
	var (
		req     dashapi.PollBugsRequest
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.ReportingPollBugs(req.Type)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportingPollNotifs(c *gin.Context, client, key string) {
	var (
		req     dashapi.PollNotificationsRequest
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.ReportingPollNotifications(req.Type)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportingPollClosed(c *gin.Context, client, key string) {
	var (
		req     dashapi.PollClosedRequest
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.ReportingPollClosed(req.IDs)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) reportingUpdate(c *gin.Context, client, key string) {
	var (
		req     dashapi.BugUpdate
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.ReportingUpdate(&req)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) managerStats(c *gin.Context, client, key string) {
	var (
		req     dashapi.ManagerStatsReq
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	managerUptimeGauges.WithLabelValues(req.Name).Set(float64(req.UpTime))
	managerCorpusGauges.WithLabelValues(req.Name).Set(float64(req.Corpus))
	managerPCsGauges.WithLabelValues(req.Name).Set(float64(req.PCs))
	managerCoverageGauges.WithLabelValues(req.Name).Set(float64(req.Cover))
	managerCrashesCounters.WithLabelValues(req.Name).Add(float64(req.Crashes))
	managerExecsCounters.WithLabelValues(req.Name).Add(float64(req.Execs))
	managerSuppCrashesCounters.WithLabelValues(req.Name).Add(float64(req.SuppressedCrashes))
	managerFuzzingDurCounters.WithLabelValues(req.Name).Add(float64(req.FuzzingTime))

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		err := dash.UploadManagerStats(&req)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}

func (p *proxy) bugList(c *gin.Context, client, key string) {
	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.BugList()
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}

	}
	p.dashMu.RUnlock()
}

func (p *proxy) loadBug(c *gin.Context, client, key string) {
	var (
		req     dashapi.LoadBugReq
		payload = c.PostForm("payload")
		buf     = bytes.NewBufferString(payload)
	)
	r, err := gzip.NewReader(buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}
	d := json.NewDecoder(r)

	if err := d.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	if err := r.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
		return
	}

	p.dashMu.RLock()
	for _, dash := range p.dashes {
		_, err := dash.LoadBug(req.ID)
		if err != nil {
			p.dashMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": errUnknownMethod.Error()})
			return
		}
	}
	p.dashMu.RUnlock()
}
