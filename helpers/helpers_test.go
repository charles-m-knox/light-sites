package helpers

import (
	"lightsites/constants"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestGetNodeOfType(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// note: ExpectsNilNode is not really leveraged here, since
	// html.Parse("") seems to always return a properly formatted HTMl document
	// including body and HTML tags.
	tests := []struct {
		InputHTML             string
		ExpectsNilNode        bool
		ExpectedNodeAttribute string
		ExpectedNodeValue     string
	}{
		{`<body test="123"></body>`, false, "test", "123"},
		{``, false, "", ""},
	}

	for _, test := range tests {
		inputHTMLNodes, err := html.Parse(strings.NewReader(test.InputHTML))
		require.NoError(err)
		actual := GetNodeOfType(inputHTMLNodes, constants.BodyNode)
		if test.ExpectsNilNode {
			assert.Nil(actual)
		} else {
			for _, attribute := range actual.Attr {
				switch attribute.Key {
				case test.ExpectedNodeAttribute:
					assert.Equal(test.ExpectedNodeValue, attribute.Val)
				default:
					t.Error("failed to find an attribute to test, to fix this add something like <body test=\"123\"></body> tag as part of the test case, and set the test case's ExpectedNodeAttribute/Value to test and 123 respectively")
				}
			}
		}
	}
}

func TestGetTitleURLFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"'''@@)(*F)(*)(*#)(@)(*@#$%&)(@#*$",
			"f",
		},
		{
			"ugly text here!!!??",
			"ugly-text-here",
		},
		{
			"hyphenated-but - still-good!",
			"hyphenated-but-still-good",
		},
		{
			"hyphen at the end should go away and this should be 36 characters)(SD&*-",
			"hyphen-at-the-end-should-go-away-and",
		},
		{
			"Wow! This message's getting some _good_ publicity!",
			"wow-this-messages-getting-some-good",
		},
	}

	for _, test := range tests {
		actual := GetTitleURLFromString(test.input, 36, true)
		assert.Equal(t, test.expected, actual)
	}
}

func TestRenderNode(t *testing.T) {
	assert := assert.New(t)

	divNode := &html.Node{
		Type: html.ElementNode,
		Data: constants.DivNode,
	}

	tests := []struct {
		TestName   string
		InputNode  *html.Node
		OutputHTML string
	}{
		{
			"RenderNode happy path",
			divNode,
			"<div></div>",
		},
	}

	for _, test := range tests {
		actual, err := RenderNode(test.InputNode)
		assert.NoError(err)
		assert.Equal(test.OutputHTML, actual, test.TestName)
	}
}
