package executor

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/v3/pkg/engine/internal/planner/physical"
	"github.com/grafana/loki/v3/pkg/engine/internal/semconv"
	"github.com/grafana/loki/v3/pkg/engine/internal/types"
	"github.com/grafana/loki/v3/pkg/util/arrowtest"
)

func TestLinefmtParser_Process(t *testing.T) {
	timestamp := time.Now().In(time.UTC).Add(time.Duration(1) * time.Second)
	schema := arrow.NewSchema(
		[]arrow.Field{
			semconv.FieldFromIdent(semconv.ColumnIdentMessage, false),
			semconv.FieldFromIdent(semconv.ColumnIdentTimestamp, false),
			semconv.FieldFromIdent(semconv.NewIdentifier("namespace", types.ColumnTypeMetadata, types.Loki.String), false),
		},
		nil, // No metadata
	)
	tests := []struct {
		name      string
		line      string
		timeVal   arrow.Timestamp
		namespace string
		lineFmt   string
		want      string
	}{
		{
			"simple replacement",
			"this is my test line of logs",
			arrow.Timestamp(timestamp.UnixNano()),
			"dev",
			"{{.namespace}}",
			"dev",
		},
		{
			"replacement plus addition",
			"this is my test line of logs",
			arrow.Timestamp(timestamp.UnixNano()),
			"dev",
			"{{.namespace}} foo",
			"dev foo",
		},
		{
			"timestamp",
			"this is my test line of logs",
			arrow.Timestamp(timestamp.UnixNano()),
			"dev",
			"{{.timestamp}} foo",
			timestamp.In(time.UTC).Format("2006-01-02T15:04:05.999999999Z") + " foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create builders for each column
			logBuilder := array.NewStringBuilder(memory.DefaultAllocator)
			tsBuilder := array.NewTimestampBuilder(memory.DefaultAllocator, &arrow.TimestampType{Unit: arrow.Nanosecond, TimeZone: "UTC"})
			namespaceBuilder := array.NewStringBuilder(memory.DefaultAllocator)

			// Append data to the builders
			logs := make([]string, 1)
			ts := make([]arrow.Timestamp, 1)
			namespaces := make([]string, 1)

			logs[0] = tt.line
			ts[0] = tt.timeVal
			namespaces[0] = tt.namespace

			tsBuilder.AppendValues(ts, nil)
			logBuilder.AppendValues(logs, nil)
			namespaceBuilder.AppendValues(namespaces, nil)

			// Build the arrays
			logArray := logBuilder.NewArray()
			tsArray := tsBuilder.NewArray()
			namespaceArray := namespaceBuilder.NewArray()
			columns := []arrow.Array{logArray, tsArray, namespaceArray}
			recordBatch := array.NewRecordBatch(schema, columns, 1)

			parser, err := NewFormatter(tt.lineFmt)
			require.NoError(t, err)
			var result = map[string]string{}
			output, _ := parser.Process(tt.line, recordBatch, result)
			require.NoError(t, err)
			require.Equal(t, tt.want, output)
			require.Equal(t, tt.want, result["message"])
		})
	}
}

func TestLinefmtParser_ErrorType(t *testing.T) {
	expr := &physical.VariadicExpr{
		Op: types.VariadicOpParseLinefmt,
		Expressions: []physical.Expression{
			&physical.ColumnExpr{Ref: semconv.ColumnIdentMessage.ColumnRef()},
			physical.NewLiteral([]string{}),
			physical.NewLiteral("{{."),
		},
	}
	e := newExpressionEvaluator()
	schema := arrow.NewSchema([]arrow.Field{
		semconv.FieldFromIdent(semconv.ColumnIdentMessage, true),
		semconv.FieldFromIdent(semconv.ColumnIdentTimestamp, true),
	}, nil)
	input := arrowtest.Rows{
		{
			semconv.ColumnIdentMessage.FQN():   "line one",
			semconv.ColumnIdentTimestamp.FQN(): time.Unix(1, 0).UTC(),
		},
	}

	record := input.Record(memory.DefaultAllocator, schema)
	col, err := e.eval(expr, record)
	require.NoError(t, err)

	arr, ok := col.(*array.Struct)
	require.True(t, ok)
	actual, err := structToRows(arr)
	require.NoError(t, err)

	require.Equal(t, types.LinefmtParserErrorType, actual[0][semconv.ColumnIdentError.FQN()])
}
