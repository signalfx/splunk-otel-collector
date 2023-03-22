package kubernetes

import (
	// Import the monitors so that they get registered
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/apiserver"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/controllermanager"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/events"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/kubeletmetrics"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/proxy"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/scheduler"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/volumes"
)
