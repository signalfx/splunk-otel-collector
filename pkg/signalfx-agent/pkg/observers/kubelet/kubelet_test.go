package kubelet

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/neotest"
	kubelet_test "github.com/signalfx/signalfx-agent/pkg/neotest/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/observers"

	. "github.com/onsi/gomega"
)

// Test_load verifies that the raw Kubelet JSON transforms into the expected Go
// struct.
func Test_load(t *testing.T) {
	podsJSON, err := ioutil.ReadFile("testdata/pods-two-running.json")
	if err != nil {
		t.Fatal("failed loading pods.json")
	}

	loadedPods := &pods{}
	neotest.LoadJSON(t, "testdata/pods-loaded.json", loadedPods)

	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *pods
		wantErr bool
	}{
		{"load failed", args{[]byte("invalid")}, nil, true},
		{"load succeeded", args{podsJSON}, loadedPods, false},
	}
	for _, tt := range tests {
		args := tt.args
		wantErr := tt.wantErr
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadJSON(args.body)
			if (err != nil) != wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, wantErr)
				return
			}
			if !reflect.DeepEqual(got, want) {
				pretty.Ldiff(t, got, want)
				t.Error("Differences detected")
			}
		})
	}
}

func TestNoPods(t *testing.T) {

	fakeKubelet := kubelet_test.NewFakeKubelet()
	fakeKubelet.Start()

	config := &Config{
		PollIntervalSeconds: 1,
		KubeletAPI: kubelet.APIConfig{
			URL: fakeKubelet.URL().String(),
		},
	}

	var kub *Observer
	var endpoints map[services.ID]services.Endpoint

	setup := func(podJSONPath string) {
		fakeKubelet.PodJSON, _ = ioutil.ReadFile(podJSONPath)

		endpoints = make(map[services.ID]services.Endpoint)

		if kub != nil {
			kub.Shutdown()
		}

		kub = &Observer{
			serviceCallbacks: &observers.ServiceCallbacks{
				Added:   func(se services.Endpoint) { endpoints[se.Core().ID] = se },
				Removed: func(se services.Endpoint) { delete(endpoints, se.Core().ID) },
			},
		}
		if err := kub.Configure(config); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("One pod without container port", func(t *testing.T) {
		RegisterTestingT(t)
		setup("testdata/pods-without-ports.json")
		Eventually(func() int { return len(endpoints) }).Should(Equal(1))
		re := endpoints[services.ID("redis-3165242388-n1vc7-2fafcdf-portless")].(*services.ContainerEndpoint)
		Expect(re.Port).To(Equal(uint16(0)))
		Expect(re.Host).To(Equal("10.2.83.18"))
	})

	t.Run("No pods at all", func(t *testing.T) {
		RegisterTestingT(t)
		setup("testdata/pods-no-pods.json")
		Consistently(func() int { return len(endpoints) }).Should(Equal(0))
	})

	t.Run("Two running pods", func(t *testing.T) {
		RegisterTestingT(t)
		setup("testdata/pods-two-running.json")
		Eventually(func() int { return len(endpoints) }).Should(Equal(5))

		re := endpoints[services.ID("redis-3165242388-n1vc7-2fafcdf-6379")].(*services.ContainerEndpoint)
		Expect(re.Port).To(Equal(uint16(6379)))
		Expect(re.Host).To(Equal("10.2.83.18"))
		Expect(re.Container.Image).To(Equal("redis:latest"))
		Expect(re.Dimensions()["kubernetes_pod_uid"]).To(Equal("2fafcdfe-f3a7-11e6-99cc-066fe1d5e5f9"))
		Expect(re.Dimensions()["kubernetes_pod_name"]).To(Equal("redis-3165242388-n1vc7"))

		re2 := endpoints[services.ID("redis-3165242388-n1vc7-2fafcdf-7379")].(*services.ContainerEndpoint)
		Expect(re2.Port).To(Equal(uint16(7379)))
		Expect(re2.Host).To(Equal("10.2.83.18"))
		Expect(re2.Container.Image).To(Equal("redis:latest"))
		Expect(re2.Dimensions()["kubernetes_pod_uid"]).To(Equal("2fafcdfe-f3a7-11e6-99cc-066fe1d5e5f9"))
	})

	t.Run("No running pods", func(t *testing.T) {
		RegisterTestingT(t)
		setup("testdata/pods-none-running.json")
		Consistently(func() int { return len(endpoints) }).Should(Equal(0))
	})

	fakeKubelet.Close()
}
