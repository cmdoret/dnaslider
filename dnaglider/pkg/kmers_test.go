package pkg

import (
	"fmt"
	"testing"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/unikmer"
)

func TestNewKmerProfile(t *testing.T) {
	tProf := NewKmerProfile(3)
	if tProf.K != 3 {
		t.Errorf("KmerProfile generated with wrong K value.")
	}
}
func TestGetSeqKmers(t *testing.T) {
	testSeq, _ := seq.NewSeq(seq.DNA, []byte("CCTA"))
	obsProf := NewKmerProfile(3)
	obsProf.GetSeqKmers(testSeq)
	for code, freq := range obsProf.Profile {
		fmt.Println(string(unikmer.Decode(code, 3)), freq)
	}
	k1, _ := unikmer.Encode([]byte("AGG"))
	k2, _ := unikmer.Encode([]byte("TAG"))
	expProf := map[uint64]float64{k1: 1.0, k2: 1.0}
	if len(obsProf.Profile) != len(expProf) {
		t.Errorf(
			"Number of kmers in profile incorrect: got %d instead of %d",
			len(obsProf.Profile),
			len(expProf),
		)
	}
	for kmer, count := range expProf {
		if obsProf.Profile[kmer] != count {
			t.Errorf(
				"Wrong count for kmer %s: got %f instead of %f",
				unikmer.Decode(kmer, obsProf.K),
				obsProf.Profile[kmer],
				count,
			)
		}
	}
}
func TestCountsToFreqs(t *testing.T) {
	testSeq, _ := seq.NewSeq(seq.DNA, []byte("CCTA"))
	obsProf := NewKmerProfile(3)
	obsProf.GetSeqKmers(testSeq)
	obsProf.CountsToFreqs()
	k1, _ := unikmer.Encode([]byte("AGG"))
	k2, _ := unikmer.Encode([]byte("TAG"))
	expProf := map[uint64]float64{k1: 0.5, k2: 0.5}
	for kmer, count := range expProf {
		if obsProf.Profile[kmer] != count {
			t.Errorf(
				"Wrong frequency for kmer %s: got %f instead of %f",
				unikmer.Decode(kmer, obsProf.K),
				obsProf.Profile[kmer],
				count,
			)
		}
	}

}
func TestKmerDist(t *testing.T) {
	testSeq, _ := seq.NewSeq(seq.DNA, []byte("CCTA"))
	testRef, _ := seq.NewSeq(seq.DNA, []byte("CCTAAA"))
	profSeq := NewKmerProfile(3)
	profRef := NewKmerProfile(3)
	profSeq.GetSeqKmers(testSeq)
	profRef.GetSeqKmers(testRef)
	profSeq.CountsToFreqs()
	profRef.CountsToFreqs()
	expDist := 0.5
	obsDist := profSeq.KmerDist(profRef)
	if expDist != obsDist {
		t.Errorf(
			"Incorrect distance between kmer profiles:  got %f instead of %f",
			obsDist,
			expDist,
		)
	}

}
