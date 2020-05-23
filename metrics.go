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

import "github.com/prometheus/client_golang/prometheus"

var (
	rpcCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Number of requests.",
		},
		[]string{"client", "method"},
	)
)

// dashapi.Build metrics
var (
	buildCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "builds_total",
			Help: "Number of builds.",
		},
		[]string{"manager", "id", "os", "arch", "vmarch"},
	)
)

// dashapi.JobPollReq metrics
var (
	jobPollCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_poll_total",
			Help: "Number of job polls.",
		},
		[]string{"manager"},
	)
)

// dashapi.JobDoneReq metrics
var (
	jobDoneCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_done_total",
			Help: "Number of jobs completed.",
		},
		[]string{"id", "manager", "build_id", "os", "arch", "vmarch"},
	)
)

// dashapi.BuildErrorReq metrics
var (
	buildErrorCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_error_total",
			Help: "Number of job build errors.",
		},
		[]string{"manager", "id", "os", "arch", "vmarch"},
	)
)

// dashapi.ManagerStatsReq metrics
var (
	managerUptimeGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manager_uptime_total",
			Help: "Manager uptime.",
		},
		[]string{"manager"},
	)
	managerCorpusGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manager_corpus_total",
			Help: "Manager corpus total.",
		},
		[]string{"manager"},
	)
	managerPCsGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manager_pcs_total",
			Help: "Manager pcs total.",
		},
		[]string{"manager"},
	)
	managerCoverageGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manager_coverage_total",
			Help: "Manager coverage total.",
		},
		[]string{"manager"},
	)
	managerCrashesCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manager_crashes_total",
			Help: "Manager crashes total.",
		},
		[]string{"manager"},
	)
	managerSuppCrashesCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manager_supp_crashes_total",
			Help: "Manager suppressed crashes total.",
		},
		[]string{"manager"},
	)
	managerExecsCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manager_execs_total",
			Help: "Manager execs total.",
		},
		[]string{"manager"},
	)
	managerFuzzingDurCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manager_fuzzing_dur_total",
			Help: "Manager fuzzing duration total.",
		},
		[]string{"manager"},
	)
)

func init() {
	prometheus.MustRegister(rpcCounters)
	prometheus.MustRegister(buildCounters)
	prometheus.MustRegister(jobPollCounters)
	prometheus.MustRegister(buildErrorCounters)
	prometheus.MustRegister(managerUptimeGauges)
	prometheus.MustRegister(managerCorpusGauges)
	prometheus.MustRegister(managerPCsGauges)
	prometheus.MustRegister(managerCoverageGauges)
	prometheus.MustRegister(managerCrashesCounters)
	prometheus.MustRegister(managerSuppCrashesCounters)
	prometheus.MustRegister(managerExecsCounters)
	prometheus.MustRegister(managerFuzzingDurCounters)
}
