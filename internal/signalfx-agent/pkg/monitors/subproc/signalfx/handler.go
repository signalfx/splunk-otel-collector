package signalfx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/mailru/easyjson"
	sfxmodel "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/ingest-protocols/protocol/signalfx"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

const messageTypeDatapointJSONList subproc.MessageType = 200
const messageTypeDatapointProtobufList subproc.MessageType = 201

type JSONHandler struct {
	Output types.Output
	Logger logrus.FieldLogger
}

var _ subproc.MessageHandler = &JSONHandler{}

func (h *JSONHandler) ProcessMessages(ctx context.Context, dataReader subproc.MessageReceiver) {
	for {
		h.Logger.Debug("Waiting for messages")
		msgType, payloadReader, err := dataReader.RecvMessage()

		// This means we are shutdown.
		if ctx.Err() != nil {
			return
		}

		h.Logger.Debugf("Got message of type %d", msgType)

		// This is usually due to the pipe being closed
		if err != nil {
			h.Logger.WithError(err).Error("Could not receive messages")
			return
		}

		if err := h.handleMessage(msgType, payloadReader); err != nil {
			h.Logger.WithError(err).Error("Could not handle message from subprocess monitor")
			continue
		}
	}
}

var buffs = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (h *JSONHandler) handleMessage(msgType subproc.MessageType, payloadReader io.Reader) error {
	switch msgType {
	case messageTypeDatapointJSONList:
		// The following is copied from github.com/signalfx/gateway
		var d JSONDatapointV2
		if err := easyjson.UnmarshalFromReader(payloadReader, &d); err != nil {
			return err
		}
		for metricType, datapoints := range d {
			if len(datapoints) > 0 {
				mt, ok := sfxmodel.MetricType_value[strings.ToUpper(metricType)]
				if !ok {
					h.Logger.Error("Unknown metric type")
					continue
				}
				out := make([]*datapoint.Datapoint, 0, len(datapoints))
				for _, jsonDatapoint := range datapoints {
					v, err := signalfx.ValueToValue(jsonDatapoint.Value)
					if err != nil {
						h.Logger.WithError(err).Error("Unable to get value for datapoint")
						continue
					}
					out = append(out, datapoint.New(jsonDatapoint.Metric, jsonDatapoint.Dimensions, v, fromMT(sfxmodel.MetricType(mt)), fromTs(jsonDatapoint.Timestamp)))
				}
				h.Output.SendDatapoints(out...)
			}
		}

	case messageTypeDatapointProtobufList:
		jeff := buffs.Get().(*bytes.Buffer)
		defer buffs.Put(jeff)
		jeff.Reset()
		if _, err := jeff.ReadFrom(payloadReader); err != nil {
			return err
		}
		var msg sfxmodel.DataPointUploadMessage
		if err := proto.Unmarshal(jeff.Bytes(), &msg); err != nil {
			return err
		}
		pbufPoints := msg.GetDatapoints()
		out := make([]*datapoint.Datapoint, 0, len(pbufPoints))
		for i := range pbufPoints {
			dp, err := signalfx.NewProtobufDataPointWithType(pbufPoints[i], sfxmodel.MetricType_GAUGE)
			if err != nil {
				h.Logger.WithError(err).Error("Unable to convert protobuf datapoint")
				continue
			}
			out = append(out, dp)
		}
		h.Output.SendDatapoints(out...)

	case subproc.MessageTypeLog:
		return h.HandleLogMessage(payloadReader)

	default:
		return fmt.Errorf("unknown message type received %d", msgType)
	}

	return nil
}

// HandleLogMessage just passes through the reader and logger to the main JSON
// implementation
func (h *JSONHandler) HandleLogMessage(logReader io.Reader) error {
	return HandleLogMessage(logReader, h.Logger)
}

// Copied from github.com/signalfx/gateway
var fromMTMap = map[sfxmodel.MetricType]datapoint.MetricType{
	sfxmodel.MetricType_CUMULATIVE_COUNTER: datapoint.Counter,
	sfxmodel.MetricType_GAUGE:              datapoint.Gauge,
	sfxmodel.MetricType_COUNTER:            datapoint.Count,
}

func fromMT(mt sfxmodel.MetricType) datapoint.MetricType {
	ret, exists := fromMTMap[mt]
	if exists {
		return ret
	}
	panic(fmt.Sprintf("Unknown metric type: %v\n", mt))
}

func fromTs(ts int64) time.Time {
	if ts > 0 {
		return time.Unix(0, ts*time.Millisecond.Nanoseconds())
	}
	return time.Now().Add(-time.Duration(time.Millisecond.Nanoseconds() * ts))
}

// LogMessage represents the log message that comes back from subprocs
type LogMessage struct {
	Message     string  `json:"message"`
	Level       string  `json:"level"`
	Logger      string  `json:"logger"`
	SourcePath  string  `json:"source_path"`
	LineNumber  int     `json:"lineno"`
	CreatedTime float64 `json:"created"`
}

// HandleLogMessage will decode a log message from the given logReader and log
// it using the provided logger.
func HandleLogMessage(logReader io.Reader, logger logrus.FieldLogger) error {
	var msg LogMessage
	err := json.NewDecoder(logReader).Decode(&msg)
	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"logger":      msg.Logger,
		"sourcePath":  msg.SourcePath,
		"lineno":      msg.LineNumber,
		"createdTime": msg.CreatedTime,
	}

	switch msg.Level {
	case "DEBUG":
		logger.WithFields(fields).Debug(msg.Message)
	case "INFO":
		logger.WithFields(fields).Info(msg.Message)
	case "WARNING":
		logger.WithFields(fields).Warn(msg.Message)
	case "ERROR":
		logger.WithFields(fields).Error(msg.Message)
	case "SEVERE":
		logger.WithFields(fields).Error(msg.Message)
	case "CRITICAL":
		logger.WithFields(fields).Errorf("CRITICAL: %s", msg.Message)
	default:
		logger.WithFields(fields).Info(msg.Message)
	}

	return nil
}
