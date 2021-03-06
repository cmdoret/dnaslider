package pkg

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"
)

var kmerRegex = regexp.MustCompile("([0-9]+)MER")

// Chunk is a piece of fastx.Record sequence, containing the associated
// id genomic coordinate. It also contains indices of sliding windows
// in which statistics will be computed.
type Chunk struct {
	ID      []byte
	BpStart int
	BpEnd   int
	wSize   int
	wStride int
	Seq     *seq.Seq
	Starts  []int
}

// ChunkResult stores the computed statistics of windows from a chunk in the
// form of a table. Each row is a window, each column is a feature.
type ChunkResult struct {
	Header []string
	Data   [][]string
}

// MinInt returns the smallest of two integers. If both are equal, the second
// input is returned.
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// MakeRange generates a slice of ints from start to end, where each value is
// spaced by step
func MakeRange(start, end, step int) []int {
	arr := make([]int, 0, 1+(end-start)/step)
	for start <= end {
		arr = append(arr, start)
		start += step
	}
	return arr
}

// Build2dSlice builds a 2d slice of float64 of target size
func Build2dSlice(rows int, cols int) [][]string {
	slice2d := make([][]string, rows)
	for i := range slice2d {
		slice2d[i] = make([]string, cols)
	}
	return slice2d
}

// ChunkGenome receives fastx.Record from a channel and produces a channel of record
// chunks. Each chunk contains multiple windows. The chunkSize is given in number of
// windows the windows size and stride are in basepair.
func ChunkGenome(records <-chan fastx.Record, winSize int, winStride int, chunkSize int) <-chan Chunk {
	chunkLen := winSize + (chunkSize-1)*winStride
	chunks := make(chan Chunk, 5)
	var bpStart, bpEnd, seqLen int
	go func() {
		for rec := range records {
			seqLen = len(rec.Seq.Seq)
			bpEnd = 0
			bpStart = 1
			// We do not compute truncated windows (stopping before
			// the end of chromosomes)
			for bpStart < (seqLen - winSize) {
				bpEnd = MinInt(bpStart+chunkLen-1, seqLen)
				chunk := Chunk{
					ID:      rec.ID,
					BpStart: bpStart,
					BpEnd:   bpEnd,
					Starts:  MakeRange(0, bpEnd-bpStart-(winSize-1), winStride),
					wSize:   winSize,
					wStride: winStride,
					Seq:     rec.Seq.SubSeq(bpStart, bpEnd),
				}

				chunks <- chunk
				bpStart = bpEnd + 1
			}
		}
		close(chunks)
		return
	}()
	return chunks
}

// ConsumeChunks computes window-based statistics in chunks and stores them in a ChunkResult struct.
func ConsumeChunks(chunks <-chan Chunk, metrics []string, refProfile map[int]KmerProfile, distMetric string) chan ChunkResult {
	var end int
	var stat float64
	out := make(chan ChunkResult, 5)
	// Mapping of {column indices: field} for regular metrics (GC content, skew, ...)
	metricCols := make(map[int]string)
	// Mapping of {column indices:  k-mer sizes} for kmer columns
	kmerCols := make(map[int]int)
	// There are 3 columns for coordinates (chrom start end), and 1 per feature
	nFeatures := 3 + len(metrics)
	// Generate column names
	header := []string{"chrom", "start", "end"}
	header = append(header, metrics...)
	// Identifying which columns contain k-mer metrics
	for idx, field := range metrics {
		if kmerRegex.Match([]byte(field)) {
			k, err := strconv.Atoi(kmerRegex.FindStringSubmatch(field)[1])
			checkError(err)
			kmerCols[idx+3] = k
		} else {
			metricCols[idx+3] = field
		}
	}
	go func() {
		for chunk := range chunks {
			nWindows := len(chunk.Starts)
			results := ChunkResult{header, Build2dSlice(nWindows, nFeatures)}
			for winID, start := range chunk.Starts {
				end = MinInt(chunk.BpStart+start+chunk.wSize-1, chunk.BpEnd)
				results.Data[winID][0] = fmt.Sprint(string(chunk.ID))
				results.Data[winID][1] = fmt.Sprint(chunk.BpStart + start)
				results.Data[winID][2] = fmt.Sprint(end)
				winSeq := chunk.Seq.SubSeq(start+1, start+chunk.wSize)
				// Compute standard fields via dispatcher map
				for colNum, field := range metricCols {
					stat = fieldDispatcher[field](winSeq)
					results.Data[winID][colNum] = fmt.Sprintf("%f", stat)
				}
				// Compute k-mer fields via dedicated function
				for colNum, k := range kmerCols {
					stat = SeqKmerDiv(winSeq, refProfile[k], distMetric)
					results.Data[winID][colNum] = fmt.Sprintf("%f", stat)
				}
			}
			out <- results
		}
		close(out)
		return
	}()
	return out
}
