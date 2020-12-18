package helpers

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFileDirectory(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := []struct {
		InputDirectory string
		ExpectedResult []string
	}{
		{"../tests/testdir1", []string{"test1", "test2", "test3"}},
		{"../tests/testdir2", []string(nil)},
	}

	for _, test := range tests {
		// load files from the specified input directory first
		actualFiles, err := ReadDirectory(test.InputDirectory)
		require.NoError(err)

		actualFileNames := LoadFileDirectory(actualFiles)
		sort.Strings(actualFileNames)
		// pipe the actualFiles into the function to be tested
		assert.Equal(test.ExpectedResult, actualFileNames)
	}
}

func TestReadDirectory(t *testing.T) {
	assert := assert.New(t)

	// don't need to test the []os.FileInfo value at this time,
	// since that's covered (accidentally) by another test.
	// These tests just need to validate error handling
	tests := []struct {
		InputDirectory string
		ExpectsError   bool
	}{
		{"/does-not-exist", true},
	}

	for _, test := range tests {
		_, err := ReadDirectory(test.InputDirectory)
		if test.ExpectsError {
			assert.Error(err)
		}
	}
}

func TestWalkDirectory(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		InputDirList    DirectoryListing
		ExpectedDirList DirectoryListing
		ExpectsError    bool
	}{
		{
			DirectoryListing{
				Path:  "../tests/walkstep",
				Files: []string{},
			},
			DirectoryListing{
				Path: "../tests/walkstep",
				Files: []string{
					"nested1/nestedtest1",
					"nested2/nestedtest2",
					"test1",
				},
			},
			false,
		},
		{
			DirectoryListing{
				Path:  "../tests/does-not-exist",
				Files: []string{},
			},
			DirectoryListing{},
			true,
		},
	}

	for _, test := range tests {
		err := test.InputDirList.WalkDirectory()
		if test.ExpectsError {
			assert.Error(err)
		} else {
			assert.NoError(err)
			assert.Equal(test.ExpectedDirList, test.InputDirList)
		}
	}
}
