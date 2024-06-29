package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// MAXIMUM_OPERATIONS_COUNT 是最大操作数
const MAXIMUM_OPERATIONS_COUNT = 64

// PerformanceCounter 表示性能计数器
type PerformanceCounter struct {
	counters []*Counter
}

// NewPerformanceCounter 创建新的性能计数器
func NewPerformanceCounter() *PerformanceCounter {
	pc := &PerformanceCounter{
		counters: make([]*Counter, MAXIMUM_OPERATIONS_COUNT),
	}
	for i := 0; i < MAXIMUM_OPERATIONS_COUNT; i += 1 {
		pc.counters[i] = NewCounter()
	}
	return pc
}

// Begin 开始计时
func (pc *PerformanceCounter) Begin(operationID int) {
	if operationID < 0 || operationID >= MAXIMUM_OPERATIONS_COUNT {
		panic(fmt.Sprintf("Invalid operation ID: %d", operationID))
	}

	counter := pc.counters[operationID]
	if !counter.beginTime.IsZero() {
		fmt.Printf("Counter (op_id: %d) already began, please call end first.\n", operationID)
		return
	}
	counter.beginTime = time.Now()
}

// End 结束计时
func (pc *PerformanceCounter) End(operationID int) {
	if operationID < 0 || operationID >= MAXIMUM_OPERATIONS_COUNT {
		panic(fmt.Sprintf("Invalid operation ID: %d", operationID))
	}

	counter := pc.counters[operationID]
	if counter.beginTime.IsZero() {
		fmt.Printf("Counter (op_id: %d) not began, please call begin first.\n", operationID)
		return
	}

	elapsed := time.Now().Sub(counter.beginTime)

	counter.totalElapsed += elapsed
	counter.opsCount += 1

	if counter.minimalElapsed == -1 || elapsed < counter.minimalElapsed {
		counter.minimalElapsed = elapsed
	}

	if elapsed > counter.maximumElapsed {
		counter.maximumElapsed = elapsed
	}

	counter.beginTime = time.Time{}
}

// Reset 重置计数器
func (pc *PerformanceCounter) Reset(operationID int) {
	if operationID < 0 || operationID >= MAXIMUM_OPERATIONS_COUNT {
		panic(fmt.Sprintf("Invalid operation ID: %d", operationID))
	}
	pc.counters[operationID].reset()
}

// ResetAll 重置所有计数器
func (pc *PerformanceCounter) ResetAll() {
	for i := range pc.counters {
		pc.counters[i].reset()
	}
}

// Total 返回总时间
func (pc *PerformanceCounter) Total(operationID int) time.Duration {
	return pc.counters[operationID].totalElapsed
}

// OpsCount 返回操作次数
func (pc *PerformanceCounter) OpsCount(operationID int) int64 {
	return pc.counters[operationID].opsCount
}

// Minimal 返回最小时间
func (pc *PerformanceCounter) Minimal(operationID int) time.Duration {
	return pc.counters[operationID].minimalElapsed
}

// Maximum 返回最大时间
func (pc *PerformanceCounter) Maximum(operationID int) time.Duration {
	return pc.counters[operationID].maximumElapsed
}

// Average 返回平均时间
func (pc *PerformanceCounter) Average(operationID int) time.Duration {
	return pc.counters[operationID].average()
}

// DumpTableAsString 以字符串形式输出性能统计结果
func (p *PerformanceCounter) DumpTableAsString(operation_names []string) string {
	var table strings.Builder

	op_cnt := len(operation_names)
	if op_cnt == 0 {
		return table.String()
	}
	if op_cnt > MAXIMUM_OPERATIONS_COUNT {
		panic(fmt.Sprintf("Invalid operation count: %d", op_cnt))
	}

	var max_rec Counter
	var max_avg time.Duration
	for _, c := range p.counters {
		if c.totalElapsed > max_rec.totalElapsed {
			max_rec.totalElapsed = c.totalElapsed
		}

		if c.opsCount > max_rec.opsCount {
			max_rec.opsCount = c.opsCount
		}

		if c.minimalElapsed > max_rec.minimalElapsed {
			max_rec.minimalElapsed = c.minimalElapsed
		}

		if c.maximumElapsed > max_rec.maximumElapsed {
			max_rec.maximumElapsed = c.maximumElapsed
		}

		if c.average() > max_avg {
			max_avg = c.average()
		}
	}

	hdr_labels := []string{"Operation", "all (us)", "ops.(times)", "min.(us)", "max.(us)", "avg.(us)"}
	col_width := make([]int, 6)

	max_name_len := 0
	for _, s := range operation_names {
		if len(s) > max_name_len {
			max_name_len = len(s)
		}
	}
	col_width[0] = int(math.Max(float64(max_name_len), float64(len(hdr_labels[0]))))

	calc_digit := func(v int64) int {
		d := 0
		for v != 0 {
			v /= 10
			d++
		}
		if v == 0 {
			return 1
		}
		return d
	}

	col_width[1] = int(math.Max(float64(calc_digit(max_rec.totalElapsed.Microseconds())), float64(len(hdr_labels[1]))))
	col_width[2] = int(math.Max(float64(calc_digit(max_rec.opsCount)), float64(len(hdr_labels[2]))))
	col_width[3] = int(math.Max(float64(calc_digit(max_rec.minimalElapsed.Microseconds())), float64(len(hdr_labels[3]))))
	col_width[4] = int(math.Max(float64(calc_digit(max_rec.maximumElapsed.Microseconds())), float64(len(hdr_labels[4]))))
	col_width[5] = int(math.Max(float64(calc_digit(max_avg.Microseconds())), float64(len(hdr_labels[5]))))

	var seg_line strings.Builder
	for _, width := range col_width {
		seg_line.WriteString("+")
		seg_line.WriteString(strings.Repeat("-", width+2))
	}
	seg_line.WriteString("+\n")

	table.WriteString(seg_line.String())

	for i, label := range hdr_labels {
		line_buf := fmt.Sprintf("| %*s ", col_width[i], label)
		table.WriteString(line_buf)
	}
	table.WriteString("|\n")
	table.WriteString(seg_line.String())

	for row, name := range operation_names {
		line_buf := fmt.Sprintf("| %*s | %*d | %*d | %*d | %*d | %*d |\n",
			col_width[0], name,
			col_width[1], p.Total(row).Microseconds(),
			col_width[2], p.OpsCount(row),
			col_width[3], p.Minimal(row).Microseconds(),
			col_width[4], p.Maximum(row).Microseconds(),
			col_width[5], p.Average(row).Microseconds())
		table.WriteString(line_buf)
	}

	table.WriteString(seg_line.String())

	return table.String()
}

type Counter struct {
	totalElapsed   time.Duration
	opsCount       int64
	minimalElapsed time.Duration
	maximumElapsed time.Duration
	beginTime      time.Time
}

func NewCounter() *Counter {
	c := &Counter{}
	c.reset()
	return c
}

func (c *Counter) reset() {
	c.totalElapsed = 0
	c.opsCount = 0
	c.maximumElapsed = 0
	c.minimalElapsed = -1
	c.beginTime = time.Time{}
}

func (c *Counter) average() time.Duration {
	if c.opsCount > 0 {
		return c.totalElapsed / time.Duration(c.opsCount)
	} else {
		return 0
	}
}
