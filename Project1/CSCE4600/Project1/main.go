package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func main() {
	// CLI args
	f, closeFile, err := openProcessingFile(os.Args...)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()

	// Load and parse processes
	processes, err := loadProcesses(f)
	if err != nil {
		log.Fatal(err)
	}

	// First-come, first-serve scheduling
	// FCFSSchedule(os.Stdout, "First-come, first-serve", processes)

	SJFSchedule(os.Stdout, "Shortest-job-first", processes)
	//
	SJFPrioritySchedule(os.Stdout, "SJFPrioritySchedule", processes)
	//
	RRSchedule(os.Stdout, "Round-robin", processes, 2)
}

func openProcessingFile(args ...string) (*os.File, func(), error) {
	if len(args) != 2 {
		return nil, nil, fmt.Errorf("%w: must give a scheduling file to process", ErrInvalidArgs)
	}
	// Read in CSV process CSV file
	f, err := os.Open(args[1])
	if err != nil {
		return nil, nil, fmt.Errorf("%v: error opening scheduling file", err)
	}
	closeFn := func() {
		if err := f.Close(); err != nil {
			log.Fatalf("%v: error closing scheduling file", err)
		}
	}

	return f, closeFn, nil
}

type (
	Process struct {
		ProcessID     int64
		ArrivalTime   int64
		BurstDuration int64
		Priority      int64
	}
	TimeSlice struct {
		PID   int64
		Start int64
		Stop  int64
	}
)

//region Schedulers

// FCFSSchedule outputs a schedule of processes in a GANTT chart and a table of timing given:
// • an output writer
// • a title for the chart
// • a slice of processes
func FCFSSchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		waitingTime     int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)
	for i := range processes {
		if processes[i].ArrivalTime > 0 {
			waitingTime = serviceTime - processes[i].ArrivalTime
		}
		totalWait += float64(waitingTime)

		start := waitingTime + processes[i].ArrivalTime

		turnaround := processes[i].BurstDuration + waitingTime
		totalTurnaround += float64(turnaround)

		completion := processes[i].BurstDuration + processes[i].ArrivalTime + waitingTime
		lastCompletion = float64(completion)

		schedule[i] = []string{
			fmt.Sprint(processes[i].ProcessID),
			fmt.Sprint(processes[i].Priority),
			fmt.Sprint(processes[i].BurstDuration),
			fmt.Sprint(processes[i].ArrivalTime),
			fmt.Sprint(waitingTime),
			fmt.Sprint(turnaround),
			fmt.Sprint(completion),
		}
		serviceTime += processes[i].BurstDuration

		gantt = append(gantt, TimeSlice{
			PID:   processes[i].ProcessID,
			Start: start,
			Stop:  serviceTime,
		})
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime int64
		totalWait   int64
		arrived     = make([]Process, 0, len(processes))
		waiting     = make(PriorityQueue, 0, len(processes))
		schedule    = make([][]string, len(processes))
		gantt       = make([]TimeSlice, 0)
	)

	for i := range processes {
		arrived = append(arrived, processes[i])
	}

	addToWaitingQueue(&waiting, &arrived, (arrived[0].ArrivalTime))

	for len(waiting) > 0 || len(arrived) > 0 {
		if len(waiting) == 0 {
			nextArrival := (arrived[0].ArrivalTime)
			addToWaitingQueue(&waiting, &arrived, nextArrival)
			continue
		}

		currentProcess := waiting.dequeue()
		serviceStart := int64(math.Max(float64(serviceTime), float64(currentProcess.ArrivalTime)))

		schedule[currentProcess.ProcessID-1] = []string{
			fmt.Sprint(currentProcess.ProcessID),
			fmt.Sprint(currentProcess.Priority),
			fmt.Sprint(currentProcess.BurstDuration),
			fmt.Sprint(currentProcess.ArrivalTime),
			fmt.Sprint(serviceStart - (currentProcess.ArrivalTime)),
			fmt.Sprint(math.Round(float64(serviceStart)-float64(currentProcess.ArrivalTime)) + float64(currentProcess.BurstDuration)),
			fmt.Sprint((serviceStart) - (currentProcess.ArrivalTime) + currentProcess.BurstDuration + currentProcess.ArrivalTime),
		}

		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: int64(serviceStart),
			Stop:  serviceStart + int64(currentProcess.BurstDuration),
		})

		totalWait += serviceStart - (currentProcess.ArrivalTime)
		serviceTime = serviceStart + (currentProcess.BurstDuration)

		if len(arrived) > 0 && serviceTime >= (arrived[0].ArrivalTime) {
			addToWaitingQueue(&waiting, &arrived, serviceTime)
		}
	}

	count := float64(len(processes))
	aveWait := float64(totalWait) / count
	aveTurnaround := float64(totalWait+serviceTime) / count
	aveThroughput := count / float64(serviceTime)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

type PriorityQueue []Process

// New function to sort Process struct by Burst Duration and Priority (in descending order)
func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].BurstDuration < pq[j].BurstDuration {
		return true
	} else if pq[i].BurstDuration == pq[j].BurstDuration {
		return pq[i].Priority > pq[j].Priority
	}
	return false
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Enqueue adds a process to the priority queue
func (pq *PriorityQueue) enqueue(p Process) {
	*pq = append(*pq, p)
	sort.Sort(pq)
}

// Dequeue removes the highest priority process from the queue
func (pq *PriorityQueue) dequeue() Process {
	p := (*pq)[0]
	*pq = (*pq)[1:]
	return p
}

// addToWaitingQueue adds arrived processes to priority queue when scheduled time has arrived
func addToWaitingQueue(waiting *PriorityQueue, arrived *[]Process, serviceTime int64) {
	for _, p := range *arrived {
		if (p.ArrivalTime) <= serviceTime {
			*waiting = append(*waiting, p)
		}
	}
	sort.Sort(waiting)
	for len(*arrived) > 0 && ((*arrived)[0].ArrivalTime) <= serviceTime {
		p := (*arrived)[0]
		*arrived = (*arrived)[1:]
		*waiting = append(*waiting, p)
	}
}

//
// func SJFSchedule(w io.Writer, title string, processes []Process) {

// }
func SJFSchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime int64
		totalWait   int64
		arrived     = make([]Process, 0, len(processes))
		waiting     = make(PriorityQueue, 0, len(processes))
		schedule    = make([][]string, len(processes))
		gantt       = make([]TimeSlice, 0)
	)

	for i := range processes {
		arrived = append(arrived, processes[i])
	}

	addToWaitingQueue(&waiting, &arrived, (arrived[0].ArrivalTime))

	for len(waiting) > 0 || len(arrived) > 0 {
		if len(waiting) == 0 {
			nextArrival := (arrived[0].ArrivalTime)
			addToWaitingQueue(&waiting, &arrived, nextArrival)
			continue
		}

		// Select the process with the shortest remaining time
		nextProcess := waiting[0]
		for _, p := range waiting[1:] {
			if p.BurstDuration < nextProcess.BurstDuration {
				nextProcess = p
			}
		}

		if nextProcess.ArrivalTime > serviceTime {
			// No process is available to run
			serviceTime = nextProcess.ArrivalTime
		}

		// Execute the selected process for one time unit
		serviceStart := serviceTime
		serviceTime++
		nextProcess.BurstDuration--

		if nextProcess.BurstDuration == 0 {
			// The process has completed
			totalWait += serviceTime - nextProcess.ArrivalTime - nextProcess.BurstDuration
			schedule[nextProcess.ProcessID-1] = []string{
				fmt.Sprint(nextProcess.ProcessID),
				fmt.Sprint(nextProcess.Priority),
				fmt.Sprint(nextProcess.BurstDuration + 1),
				fmt.Sprint(nextProcess.ArrivalTime),
				fmt.Sprint(serviceStart - nextProcess.ArrivalTime),
				fmt.Sprint(serviceTime - nextProcess.ArrivalTime),
				fmt.Sprint(serviceTime),
			}
		} else {
			// The process is not completed yet, add it back to the waiting queue
			waiting = waiting[1:]
			addToWaitingQueue(&waiting, &arrived, serviceTime)
		}

		// Add any new arrived processes to the waiting queue
		addToWaitingQueue(&waiting, &arrived, serviceTime)

		gantt = append(gantt, TimeSlice{
			PID:   nextProcess.ProcessID,
			Start: serviceStart,
			Stop:  serviceTime,
		})
	}

	count := float64(len(processes))
	aveWait := float64(totalWait) / count
	aveTurnaround := float64(totalWait+serviceTime) / count
	// aveTurnaround := aveWait + float64(schedule[len(schedule)-1][6])
	aveThroughput := count / float64(serviceTime)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func RRSchedule(w io.Writer, title string, processes []Process, quantum int64) {
	var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		waiting         = make([]Process, 0, len(processes))
		schedule        = make([][]string, 0, len(processes))
		gantt           = make([]TimeSlice, 0)
	)

	for i := range processes {
		waiting = append(waiting, processes[i])
	}

	for len(waiting) > 0 {
		p := waiting[0]
		waiting = waiting[1:]

		start := int64(math.Max(float64(serviceTime), float64(p.ArrivalTime)))
		duration := int64(math.Min(float64(p.BurstDuration), float64(quantum)))

		serviceTime += duration
		p.BurstDuration -= duration

		if p.BurstDuration > 0 {
			waiting = append(waiting, p)
		} else {
			turnaround := serviceTime - p.ArrivalTime
			totalTurnaround += float64(turnaround)
			waitingTime := start - p.ArrivalTime
			totalWait += float64(waitingTime)

			schedule = append(schedule, []string{
				fmt.Sprint(p.ProcessID),
				fmt.Sprint(p.Priority),
				fmt.Sprint(p.BurstDuration + duration),
				fmt.Sprint(p.ArrivalTime),
				fmt.Sprint(waitingTime),
				fmt.Sprint(turnaround),
				fmt.Sprint(serviceTime),
			})
		}

		gantt = append(gantt, TimeSlice{
			PID:   p.ProcessID,
			Start: start,
			Stop:  serviceTime,
		})
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / float64(serviceTime)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

//endregion

//region Output helpers

func outputTitle(w io.Writer, title string) {
	_, _ = fmt.Fprintln(w, strings.Repeat("-", len(title)*2))
	_, _ = fmt.Fprintln(w, strings.Repeat(" ", len(title)/2), title)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", len(title)*2))
}

func outputGantt(w io.Writer, gantt []TimeSlice) {
	_, _ = fmt.Fprintln(w, "Gantt schedule")
	_, _ = fmt.Fprint(w, "|")
	for i := range gantt {
		pid := fmt.Sprint(gantt[i].PID)
		padding := strings.Repeat(" ", (8-len(pid))/2)
		_, _ = fmt.Fprint(w, padding, pid, padding, "|")
	}
	_, _ = fmt.Fprintln(w)
	for i := range gantt {
		_, _ = fmt.Fprint(w, fmt.Sprint(gantt[i].Start), "\t")
		if len(gantt)-1 == i {
			_, _ = fmt.Fprint(w, fmt.Sprint(gantt[i].Stop))
		}
	}
	_, _ = fmt.Fprintf(w, "\n\n")
}

func outputSchedule(w io.Writer, rows [][]string, wait, turnaround, throughput float64) {
	_, _ = fmt.Fprintln(w, "Schedule table")
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"ID", "Priority", "Burst", "Arrival", "Wait", "Turnaround", "Exit"})
	table.AppendBulk(rows)
	table.SetFooter([]string{"", "", "", "",
		fmt.Sprintf("Average\n%.2f", wait),
		fmt.Sprintf("Average\n%.2f", turnaround),
		fmt.Sprintf("Throughput\n%.2f/t", throughput)})
	table.Render()
}

//endregion

//region Loading processes.

var ErrInvalidArgs = errors.New("invalid args")

func loadProcesses(r io.Reader) ([]Process, error) {
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: reading CSV", err)
	}

	processes := make([]Process, len(rows))
	for i := range rows {
		processes[i].ProcessID = mustStrToInt(rows[i][0])
		processes[i].BurstDuration = mustStrToInt(rows[i][1])
		processes[i].ArrivalTime = mustStrToInt(rows[i][2])
		if len(rows[i]) == 4 {
			processes[i].Priority = mustStrToInt(rows[i][3])
		}
	}

	return processes, nil
}

func mustStrToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return i
}

//endregion
