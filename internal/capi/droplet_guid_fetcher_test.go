package capi_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/poy/cf-faas/internal/capi"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

type TD struct {
	*testing.T
	spyDropletClient *spyDropletClient
	f                *capi.DropletGuidFetcher
}

func TestDropletGuidFetcher(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TD {
		spyDropletClient := newSpyDropletClient()
		return TD{
			T:                t,
			spyDropletClient: spyDropletClient,
			f:                capi.NewDropletGuidFetcher(spyDropletClient),
		}
	})

	o.Spec("it asks for the name, the droplet guid and then copies", func(t TD) {
		t.spyDropletClient.appResult = "some-guid"
		t.spyDropletClient.dropletResult = "droplet-guid"
		ctx, _ := context.WithCancel(context.Background())

		appGuid, dropletGuid, err := t.f.FetchGuid(ctx, "some-name")
		Expect(t, err).To(BeNil())

		Expect(t, appGuid).To(Equal("some-guid"))
		Expect(t, dropletGuid).To(Equal("droplet-guid"))

		Expect(t, t.spyDropletClient.appCtx).To(Equal(ctx))
		Expect(t, t.spyDropletClient.appName).To(Equal("some-name"))

		Expect(t, t.spyDropletClient.dropletCtx).To(Equal(ctx))
		Expect(t, t.spyDropletClient.appGuid).To(Equal("some-guid"))
	})

	o.Spec("it only copies new droplets", func(t TD) {
		t.spyDropletClient.appResult = "some-guid"
		t.spyDropletClient.dropletResult = "droplet-guid"
		ctx, _ := context.WithCancel(context.Background())

		t.f.FetchGuid(ctx, "some-name")
		t.f.FetchGuid(ctx, "some-name")
	})

	o.Spec("it returns an error if fetching the app guid fails", func(t TD) {
		t.spyDropletClient.appResult = "some-guid"
		t.spyDropletClient.dropletErr = errors.New("some-error")
		_, _, err := t.f.FetchGuid(context.Background(), "some-name")
		Expect(t, err).To(Not(BeNil()))
	})

	o.Spec("it returns an error if fetching the droplet guid fails", func(t TD) {
		t.spyDropletClient.dropletErr = errors.New("some-error")
		_, _, err := t.f.FetchGuid(context.Background(), "some-name")
		Expect(t, err).To(Not(BeNil()))
	})

	o.Spec("it returns an empty string for an empty app name", func(t TD) {
		appGuid, dropletGuid, err := t.f.FetchGuid(context.Background(), "")
		Expect(t, err).To(BeNil())
		Expect(t, appGuid).To(HaveLen(0))
		Expect(t, dropletGuid).To(HaveLen(0))
		Expect(t, t.spyDropletClient.appCtx).To(BeNil())
	})

	o.Spec("it survives the race detector", func(t TD) {
		go func() {
			for i := 0; i < 100; i++ {
				t.f.FetchGuid(context.Background(), "some-name")
			}
		}()

		for i := 0; i < 100; i++ {
			t.f.FetchGuid(context.Background(), "some-name")
		}
	})
}

type spyDropletClient struct {
	mu        sync.Mutex
	appCtx    context.Context
	appName   string
	appResult string
	appErr    error

	dropletCtx    context.Context
	appGuid       string
	dropletResult string
	dropletErr    error
}

func newSpyDropletClient() *spyDropletClient {
	return &spyDropletClient{}
}

func (s *spyDropletClient) GetAppGuid(ctx context.Context, appName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appCtx = ctx
	s.appName = appName
	return s.appResult, s.appErr
}

func (s *spyDropletClient) GetDropletGuid(ctx context.Context, appGuid string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropletCtx = ctx
	s.appGuid = appGuid
	return s.dropletResult, s.dropletErr
}
