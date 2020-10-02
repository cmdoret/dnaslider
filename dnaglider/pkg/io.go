package pkg

import (
	"io"
	"log"
	"os"

	"github.com/shenwei356/bio/seqio/fastx"
)

// StreamGenome reads records one by one from an input fasta file and sends them
// to a channel for downstream processing
func StreamGenome(fasta string, bufSize int) <-chan fastx.Record {
	var record *fastx.Record
	recordChan := make(chan fastx.Record, bufSize)
	reader, err := fastx.NewDefaultReader(fasta)
	// Can't read input path
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	// Read records asynchronously and send them to a channel
	go func() {
		for {
			record, err = reader.Read()
			if err != nil {
				// Reached last record
				if err == io.EOF {
					close(recordChan)
					return
				}
				log.Fatal(err)
				break
			}
			recordChan <- *record.Clone()
		}
		close(recordChan)
	}()
	return recordChan

}

// FastaToKmers reads all records in a fasta file and computes its k-mer profile
func FastaToKmers(fasta string, k int) KmerProfile {
	var record *fastx.Record
	var profile = KmerProfile{k, make(map[string]float64)}
	reader, err := fastx.NewDefaultReader(fasta)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	for {
		record, err = reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
			break
		}
		// Add K-mer counts for each record
		profile.GetSeqKmers(record.Seq)
	}
	// Once all records have been computed, convert to freqs
	profile.CountsToFreqs()
	return profile
}