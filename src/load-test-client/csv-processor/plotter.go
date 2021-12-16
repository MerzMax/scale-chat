package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"os"
)

func Plot(fileName string, entries *[]MessageLatencyEntry) {
	page := components.NewPage()

	xValues := generateXValuesTimeSeries(entries)

	page.AddCharts(
		plotRtts(fileName, entries, &xValues),
		plotAverageLatency(fileName, entries, &xValues),
	)

	f, err := os.Create(outputDir + fileName + "_line-rtt.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))
}

func generateXValuesTimeSeries(entries *[]MessageLatencyEntry) []string {
	values := make([]string, len(*entries))
	for i, entry := range *entries {
		values[i] = entry.SenderMsgEvent.TimeStamp.Format("15:04:105")
	}

	return values
}

// RTT

func plotRtts(fileName string, entries *[]MessageLatencyEntry, xValues *[]string) *charts.Line {
	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "The RTT of all messages (" + fileName + ")",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time[date]",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Rtt[ns]",
			AxisLabel: &opts.AxisLabel{
				Show: true,
			},
		}))

	line.SetXAxis(xValues).
		AddSeries("RTT", generateYValuesRtt(entries)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	return line
}

func generateYValuesRtt(entries *[]MessageLatencyEntry) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {
		values[i] = opts.LineData{
			Value: entry.RttInNs,
		}
	}

	return values
}

// AVERAGE LATENCY

func plotAverageLatency(fileName string, entries *[]MessageLatencyEntry, xValues *[]string) *charts.Line {
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "The average latency of all messages (" + fileName + ")",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time[date]",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Latency[ns]",
			AxisLabel: &opts.AxisLabel{
				Show: true,
			},
		}))

	// Put data into instance
	line.SetXAxis(xValues).
		AddSeries("average latency", generateYValuesAvgLatency(entries)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	return line
}

func generateYValuesAvgLatency(entries *[]MessageLatencyEntry) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {

		var total int64

		for _, latency := range entry.LatenciesInNs {
			total += latency
		}

		values[i] = opts.LineData{
			Value: total / int64(len(entry.LatenciesInNs)),
		}
	}

	return values
}
