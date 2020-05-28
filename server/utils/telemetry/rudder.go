package telemetry

import rudder "github.com/rudderlabs/analytics-go"

const (
	rudderDataPlaneURL = ""
	rudderWriteKey     = ""
)

func NewRudderClient() (Client, error) {
	client, err := rudder.NewWithConfig(rudderWriteKey, rudderDataPlaneURL, rudder.Config{})
	if err != nil {
		return nil, err
	}

	return &rudderWrapper{client: client}, nil
}

func NewRudderClientWithCredentials(dataPlaneURL, writeKey string) (Client, error) {
	client, err := rudder.NewWithConfig(writeKey, dataPlaneURL, rudder.Config{})
	if err != nil {
		return nil, err
	}

	return &rudderWrapper{client: client}, nil
}

type rudderWrapper struct {
	client rudder.Client
}

func (r *rudderWrapper) Enqueue(t Track) {
	r.client.Enqueue(rudder.Track{
		UserId:     t.UserID,
		Event:      t.Event,
		Properties: t.Properties,
	})
}
