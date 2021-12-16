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
	bar.SetGlobalOptions(
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

	// Put data into instance
	bar.SetXAxis(GenerateXValuesTimeSeries(entries)).
		AddSeries("RTT", GenerateYValuesRtt(entries)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	// Where the magic happens
	f, _ := os.Create(outputDir + fileName + "_line-rtt.html")
	bar.Render(f)
}

func GenerateXValuesTimeSeries(entries *[]MessageLatencyEntry) []string {
	values := make([]string, len(*entries))
	for i, entry := range *entries {
		values[i] = entry.SenderMsgEvent.TimeStamp.Format("15:04:105")
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
