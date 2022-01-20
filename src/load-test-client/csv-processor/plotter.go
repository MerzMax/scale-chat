package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/montanaflynn/stats"
	"log"
	"os"
)

func Plot(fileName string, entries *[]MessageLatencyEntry) {
	// Generate x values (identical for all charts)
	xValues := generateXValuesTimeSeries(entries)

	// Generate charts and save them in files
	rttPlot := plotRtts(fileName, entries, &xValues)
	writeLineChartToFile(rttPlot, fileName, "_line-rtt")

	latenciesPlot := plotLatency(fileName, entries, &xValues)
	writeLineChartToFile(latenciesPlot, fileName, "_line-latency")

	// Save all plots in one file
	pageWithAll := components.NewPage()
	pageWithAll.AddCharts(
		rttPlot,
		latenciesPlot,
	)
	f, err := os.Create(outputDir + fileName + "_all.html")
	if err != nil {
		panic(err)
	}
	pageWithAll.Render(f)
	log.Println("Created new chart for: " + fileName + "_all")
}

func generateXValuesTimeSeries(entries *[]MessageLatencyEntry) []string {
	values := make([]string, len(*entries))
	for i, entry := range *entries {
		values[i] = entry.SenderMsgEvent.TimeStamp.Format("15:04:105")
	}

	return values
}

func writeLineChartToFile(line *charts.Line, fileName string, suffix string) {
	f, err := os.Create(outputDir + fileName + suffix + ".html")
	if err != nil {
		log.Fatalln("Could not create file.", err)
	}
	err = line.Render(f)
	if err != nil {
		log.Fatalln("Could not render line chart to file.", err)
	}
	log.Println("Created new chart for: " + fileName + suffix)
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
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:    true,
			Padding: 40,
		}))

	line.SetXAxis(xValues).
		AddSeries("round trip time", generateYValuesRtt(entries)).
		SetSeriesOptions(
			charts.WithMarkPointNameTypeItemOpts(
				opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"},
				opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"},
			),
			charts.WithMarkPointStyleOpts(
				opts.MarkPointStyle{Label: &opts.Label{Show: true}}),
		)

	return line
}

func generateYValuesRtt(entries *[]MessageLatencyEntry) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {
		values[i] = opts.LineData{Value: entry.RttInNs}
	}

	return values
}

// AVERAGE LATENCY

func plotLatency(fileName string, entries *[]MessageLatencyEntry, xValues *[]string) *charts.Line {
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "The latency of all messages (" + fileName + ")",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time[date]",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Latency[ns]",
			AxisLabel: &opts.AxisLabel{
				Show: true,
			},
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:    true,
			Padding: 40,
		}))

	line.SetXAxis(xValues).
		AddSeries("average", generateYValuesAvgLatency(entries)).
		AddSeries("99 percentile", generateYValuesPercentilesLatency(entries, 99)).
		AddSeries("75 percentile", generateYValuesPercentilesLatency(entries, 75)).
		SetSeriesOptions(
			charts.WithMarkPointNameTypeItemOpts(
				opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"},
				opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"},
			),
			charts.WithMarkPointStyleOpts(
				opts.MarkPointStyle{Label: &opts.Label{Show: true}}),
		)

	return line
}

func generateYValuesAvgLatency(entries *[]MessageLatencyEntry) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {

		if len(entry.LatenciesInNs) == 0 {
			values[i] = opts.LineData{Value: 0}
			continue
		}

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

// PERCENTILES

func generateYValuesPercentilesLatency(entries *[]MessageLatencyEntry, percentile float64) []opts.LineData {
	values := make([]opts.LineData, len(*entries))

	for i, entry := range *entries {

		var f []float64
		for _, latency := range entry.LatenciesInNs {
			f = append(f, float64(latency))
		}

		data := stats.LoadRawData(f)
		result, err := stats.Percentile(data, percentile)
		if err != nil {
			log.Fatalln("Can't calculate percentile: ", err)
		}

		values[i] = opts.LineData{Value: fmt.Sprint(result)}
	}

	return values
}
