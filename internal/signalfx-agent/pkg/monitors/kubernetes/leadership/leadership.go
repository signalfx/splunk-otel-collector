package leadership

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	coordinationv1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

// These are global variables because all monitors share the same election
// process and it would make the agent core more complicated by trying to pass
// around an elector object to monitors that need it.
var noticeChans []chan<- bool
var lock sync.Mutex
var started bool
var isLeader bool
var leaderIdentity string

// RequestLeaderNotification provides a simple way for monitors to only send
// metrics from a single instance of the agent.  It wraps client-go's
// leaderelection tool, and uses the node name as the identifier in the
// election process, but this is scoped by namespace as well so there can be at
// most one agent pod per namespace per node for the logic to work. Calling
// this function starts the election process if it is not already running and
// returns a channel that gets fed true when this instance becomes leader and
// subsequently false if the instance stops being the leader for some reason,
// at which point the channel could send true again and so on. All monitors
// that need leader election will share the same election process.  There is no
// way to stop the leader election process once it starts.
func RequestLeaderNotification(v1Client corev1.CoreV1Interface, coordinationClient coordinationv1.CoordinationV1Interface, logger log.FieldLogger) (<-chan bool, func(), error) {
	lock.Lock()
	defer lock.Unlock()

	if !started {
		if err := startLeaderElection(v1Client, coordinationClient, logger); err != nil {
			return nil, nil, err
		}
		started = true
	}

	ch := make(chan bool, 1)

	// Prime it with the fact that we are the leader if we are -- this
	// guarantees that the first value sent to the chan will always be true.
	if isLeader {
		ch <- true
	}

	noticeChans = append(noticeChans, ch)
	return ch, func() {
		lock.Lock()
		defer lock.Unlock()

		logger.Info("Unsubscribing leader notice channel")
		for i := range noticeChans {
			if noticeChans[i] == ch {
				noticeChans = append(noticeChans[:i], noticeChans[i+1:]...)
				return
			}
		}
		logger.Error("Could not find leader notice channel to unsubscribe")
	}, nil
}

func startLeaderElection(v1Client corev1.CoreV1Interface, coordinationClient coordinationv1.CoordinationV1Interface, logger log.FieldLogger) error {
	ns := os.Getenv("MY_NAMESPACE")
	if ns == "" {
		return errors.New("MY_NAMESPACE envvar is not defined")
	}

	nodeName := os.Getenv("MY_NODE_NAME")
	if nodeName == "" {
		return errors.New("MY_NODE_NAME envvar is not defined")
	}

	resLock, err := resourcelock.New(
		resourcelock.ConfigMapsLeasesResourceLock,
		ns,
		"signalfx-agent-leader",
		v1Client,
		coordinationClient,
		resourcelock.ResourceLockConfig{
			Identity: nodeName,
			// client-go can't make anything simple
			EventRecorder: &record.FakeRecorder{},
		})

	if err != nil {
		return err
	}

	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          resLock,
		LeaseDuration: 60 * time.Second,
		RenewDeadline: 45 * time.Second,
		RetryPeriod:   30 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {},
			OnStoppedLeading: func() {},
			OnNewLeader: func(identity string) {
				lock.Lock()
				defer lock.Unlock()

				logger.Infof("K8s leader is now node %s", identity)
				leaderIdentity = identity
				if identity == nodeName && !isLeader {
					for i := range noticeChans {
						noticeChans[i] <- true
					}
					isLeader = true
				} else if identity != nodeName && isLeader {
					for i := range noticeChans {
						noticeChans[i] <- false
					}
				}
			},
		},
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			le.Run(context.Background())
		}
	}()

	return nil
}

// CurrentLeader returns the current cluster leader node, if the current agent
// instance has successfully participated in the election process and been
// notified of the leader.
func CurrentLeader() string {
	return leaderIdentity
}
