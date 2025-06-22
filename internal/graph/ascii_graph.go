package graph

import (
	"fmt"
	"math"
	"strings"

	"github.com/pterm/pterm"
	"github.com/uchr/ToDoInfo/internal/storage"
)

const (
	defaultGraphHeight = 10
	defaultGraphWidth  = 60
)

// ASCIIGraph creates ASCII graphs from time series data
type ASCIIGraph struct {
	height int
	width  int
}

// NewASCIIGraph creates a new ASCII graph renderer
func NewASCIIGraph() *ASCIIGraph {
	return &ASCIIGraph{
		height: defaultGraphHeight,
		width:  defaultGraphWidth,
	}
}

// RenderTimeSeriesGraph renders a time series graph showing max age over time
func (g *ASCIIGraph) RenderTimeSeriesGraph(points []storage.TimeSeriesPoint, title string) {
	if len(points) == 0 {
		pterm.Info.Println("No historical data available for graph")
		return
	}

	pterm.DefaultSection.Println(title)
	
	// Find min and max values
	minAge, maxAge := g.findMinMax(points)
	if maxAge == minAge {
		maxAge = minAge + 1 // Avoid division by zero
	}
	
	// Create the graph
	graph := make([][]string, g.height)
	for i := range graph {
		graph[i] = make([]string, g.width)
		for j := range graph[i] {
			graph[i][j] = " "
		}
	}
	
	// Plot the points
	for i, point := range points {
		if i >= g.width {
			break
		}
		
		// Normalize the age to graph height
		normalizedAge := float64(point.MaxAge-minAge) / float64(maxAge-minAge)
		y := int(normalizedAge * float64(g.height-1))
		y = g.height - 1 - y // Invert Y axis (0 at bottom)
		
		if y >= 0 && y < g.height && i < g.width {
			graph[y][i] = "█"
		}
	}
	
	// Render the graph with axis labels
	g.renderGraph(graph, minAge, maxAge, points)
}

// RenderTaskCountGraph renders a time series graph showing task count over time
func (g *ASCIIGraph) RenderTaskCountGraph(points []storage.TimeSeriesPoint, title string) {
	if len(points) == 0 {
		pterm.Info.Println("No historical data available for graph")
		return
	}

	pterm.DefaultSection.Println(title)
	
	// Find min and max task count values
	minCount, maxCount := g.findMinMaxTaskCount(points)
	if maxCount == minCount {
		maxCount = minCount + 1 // Avoid division by zero
	}
	
	// Create the graph
	graph := make([][]string, g.height)
	for i := range graph {
		graph[i] = make([]string, g.width)
		for j := range graph[i] {
			graph[i][j] = " "
		}
	}
	
	// Plot the points
	for i, point := range points {
		if i >= g.width {
			break
		}
		
		// Normalize the task count to graph height
		normalizedCount := float64(point.TaskCount-minCount) / float64(maxCount-minCount)
		y := int(normalizedCount * float64(g.height-1))
		y = g.height - 1 - y // Invert Y axis (0 at bottom)
		
		if y >= 0 && y < g.height && i < g.width {
			graph[y][i] = "█"
		}
	}
	
	// Render the graph with axis labels for task count
	g.renderTaskCountGraph(graph, minCount, maxCount, points)
}

// findMinMax finds the minimum and maximum ages in the data points
func (g *ASCIIGraph) findMinMax(points []storage.TimeSeriesPoint) (int, int) {
	if len(points) == 0 {
		return 0, 0
	}
	
	min := points[0].MaxAge
	max := points[0].MaxAge
	
	for _, point := range points {
		if point.MaxAge < min {
			min = point.MaxAge
		}
		if point.MaxAge > max {
			max = point.MaxAge
		}
	}
	
	return min, max
}

// findMinMaxTaskCount finds the minimum and maximum task counts in the data points
func (g *ASCIIGraph) findMinMaxTaskCount(points []storage.TimeSeriesPoint) (int, int) {
	if len(points) == 0 {
		return 0, 0
	}
	
	min := points[0].TaskCount
	max := points[0].TaskCount
	
	for _, point := range points {
		if point.TaskCount < min {
			min = point.TaskCount
		}
		if point.TaskCount > max {
			max = point.TaskCount
		}
	}
	
	return min, max
}

// renderGraph renders the graph with axes and labels
func (g *ASCIIGraph) renderGraph(graph [][]string, minAge, maxAge int, points []storage.TimeSeriesPoint) {
	// Y-axis labels (age values)
	yAxisWidth := len(fmt.Sprintf("%d", maxAge)) + 1
	
	for i := 0; i < g.height; i++ {
		// Calculate the age value for this row
		normalizedPos := float64(g.height-1-i) / float64(g.height-1)
		ageValue := int(math.Round(float64(minAge) + normalizedPos*float64(maxAge-minAge)))
		
		// Y-axis label
		yLabel := fmt.Sprintf("%*d", yAxisWidth, ageValue)
		
		// Graph row
		row := strings.Join(graph[i], "")
		
		pterm.Printf("%s│%s\n", pterm.LightBlue(yLabel), row)
	}
	
	// X-axis
	xAxisLine := strings.Repeat("─", g.width)
	xAxisPadding := strings.Repeat(" ", yAxisWidth)
	pterm.Printf("%s└%s\n", xAxisPadding, xAxisLine)
	
	// X-axis labels (dates)
	g.renderXAxisLabels(points, yAxisWidth)
	
	// Legend
	pterm.Println()
	pterm.Printf("  %s Max task age over time (days)\n", pterm.Gray("Legend:"))
	pterm.Printf("  %s Each column represents a day\n", pterm.Gray(""))
}

// renderTaskCountGraph renders the task count graph with axes and labels
func (g *ASCIIGraph) renderTaskCountGraph(graph [][]string, minCount, maxCount int, points []storage.TimeSeriesPoint) {
	// Y-axis labels (task count values)
	yAxisWidth := len(fmt.Sprintf("%d", maxCount)) + 1
	
	for i := 0; i < g.height; i++ {
		// Calculate the task count value for this row
		normalizedPos := float64(g.height-1-i) / float64(g.height-1)
		countValue := int(math.Round(float64(minCount) + normalizedPos*float64(maxCount-minCount)))
		
		// Y-axis label
		yLabel := fmt.Sprintf("%*d", yAxisWidth, countValue)
		
		// Graph row
		row := strings.Join(graph[i], "")
		
		pterm.Printf("%s│%s\n", pterm.LightBlue(yLabel), row)
	}
	
	// X-axis
	xAxisLine := strings.Repeat("─", g.width)
	xAxisPadding := strings.Repeat(" ", yAxisWidth)
	pterm.Printf("%s└%s\n", xAxisPadding, xAxisLine)
	
	// X-axis labels (dates)
	g.renderXAxisLabels(points, yAxisWidth)
	
	// Legend
	pterm.Println()
	pterm.Printf("  %s Number of tasks over time\n", pterm.Gray("Legend:"))
	pterm.Printf("  %s Each column represents a day\n", pterm.Gray(""))
}

// renderXAxisLabels renders the X-axis date labels
func (g *ASCIIGraph) renderXAxisLabels(points []storage.TimeSeriesPoint, yAxisWidth int) {
	if len(points) == 0 {
		return
	}
	
	xAxisPadding := strings.Repeat(" ", yAxisWidth+1)
	
	// Show first and last dates
	if len(points) > 0 {
		firstDate := points[0].Date.Format("Jan 02")
		lastDate := points[len(points)-1].Date.Format("Jan 02")
		
		// Create spacing for the labels
		spacing := g.width - len(firstDate) - len(lastDate)
		if spacing < 0 {
			spacing = 0
		}
		
		dateLabels := fmt.Sprintf("%s%s%s", 
			firstDate, 
			strings.Repeat(" ", spacing), 
			lastDate)
		
		pterm.Printf("%s%s\n", xAxisPadding, pterm.Gray(dateLabels))
	}
}