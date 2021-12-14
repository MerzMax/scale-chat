package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"log"
	"os"
)

func PlotRtts(fileName string, entries *[]MessageLatencyEntry) {

	log.Println("")

	// create a new bar instance
	bar := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "My first bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))

	// Put data into instance
	bar.SetXAxis(GenerateXValuesTimeSeries(entries)).
		AddSeries("Category A", GenerateYValuesRtt(entries)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	// Where the magic happens
	f, _ := os.Create(fileName + "_RTT_Line.html")
	bar.Render(f)
}

func GenerateXValuesTimeSeries(entries *[]MessageLatencyEntry) []string {
	values := make([]string, len(*entries))
	for i, entry := range *entries {
		values[i] = entry.SenderMsgEvent.TimeStamp.String()
	}

	return values
}

func GenerateYValuesRtt(entries *[]MessageLatencyEntry) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {
		values[i] = opts.LineData{
			Value: entry.RttInNs,
		}
	}

	return values
}
