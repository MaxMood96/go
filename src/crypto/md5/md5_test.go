// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package md5

import (
	"bytes"
	"crypto/internal/cryptotest"
	"crypto/rand"
	"encoding"
	"fmt"
	"hash"
	"io"
	"testing"
	"unsafe"
)

type md5Test struct {
	out       string
	in        string
	halfState string // marshaled hash state after first half of in written, used by TestGoldenMarshal
}

var golden = []md5Test{
	{"d41d8cd98f00b204e9800998ecf8427e", "", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tv\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"},
	{"0cc175b9c0f1b6a831c399e269772661", "a", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tv\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"},
	{"187ef4436122d1cc2f40dc2b92f0eba0", "ab", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tva\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01"},
	{"900150983cd24fb0d6963f7d28e17f72", "abc", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tva\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01"},
	{"e2fc714c4727ee9395f324cd2e7f331f", "abcd", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvab\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02"},
	{"ab56b4d92b40713acc5af89985d4b786", "abcde", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvab\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02"},
	{"e80b5017098950fc58aad83c8c14978e", "abcdef", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvabc\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03"},
	{"7ac66c0f148de9519b8bd264312c4d64", "abcdefg", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvabc\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03"},
	{"e8dc4081b13434b45189a720b77b6818", "abcdefgh", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvabcd\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x04"},
	{"8aa99b1f439ff71293e95357bac6fd94", "abcdefghi", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvabcd\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x04"},
	{"a925576942e94b2ef57a066101b48876", "abcdefghij", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvabcde\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05"},
	{"d747fc1719c7eacb84058196cfe56d57", "Discard medicine more than two years old.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvDiscard medicine mor\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x14"},
	{"bff2dcb37ef3a44ba43ab144768ca837", "He who has a shady past knows that nice guys finish last.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvHe who has a shady past know\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x1c"},
	{"0441015ecb54a7342d017ed1bcfdbea5", "I wouldn't marry him with a ten foot pole.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvI wouldn't marry him \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x15"},
	{"9e3cac8e9e9757a60c3ea391130d3689", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvFree! Free!/A trip/to Mars/f\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x1c"},
	{"a0f04459b031f916a59a35cc482dc039", "The days of the digital watch are numbered.  -Tom Stoppard", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvThe days of the digital watch\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x1d"},
	{"e7a48e0fe884faf31475d2a04b1362cc", "Nepal premier won't resign.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvNepal premier\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\r"},
	{"637d2fe925c07c113800509964fb0e06", "For every action there is an equal and opposite government program.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvFor every action there is an equa\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00!"},
	{"834a8d18d5c6562119cf4c7f5086cb71", "His money is twice tainted: 'taint yours and 'taint mine.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvHis money is twice tainted: \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x1c"},
	{"de3a4d2fd6c73ec2db2abad23b444281", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvThere is no reason for any individual to hav\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00,"},
	{"acf203f997e2cf74ea3aff86985aefaf", "It's a tiny change to the code and not completely disgusting. - Bob Manchek", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvIt's a tiny change to the code and no\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00%"},
	{"e1c1384cb4d2221dfdd7c795a4222c9a", "size:  a.out:  bad magic", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102Tvsize:  a.out\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\f"},
	{"c90f3ddecc54f34228c063d7525bf644", "The major problem is with sendmail.  -Mark Horton", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvThe major problem is wit\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18"},
	{"cdf7ab6c1fd49bd9933c43f3ea5af185", "Give me a rock, paper and scissors and I will move the world.  CCFestoon", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvGive me a rock, paper and scissors a\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00$"},
	{"83bc85234942fc883c063cbd7f0ad5d0", "If the enemy is within range, then so are you.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvIf the enemy is within \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x17"},
	{"277cbe255686b48dd7e8f389394d9299", "It's well we cannot hear the screams/That we create in others' dreams.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvIt's well we cannot hear the scream\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00#"},
	{"fd3fb0a7ffb8af16603f3d3af98f8e1f", "You remind me of a TV show, but that's all right: I watch it anyway.", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvYou remind me of a TV show, but th\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\""},
	{"469b13a78ebf297ecda64d4723655154", "C is as portable as Stonehedge!!", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvC is as portable\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x10"},
	{"63eb3a2f466410104731c4b037600110", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvEven if I could be Shakespeare, I think I sh\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00,"},
	{"72c2ed7592debca1c90fc0100f931a2f", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule", "md5\x01\xa7\xc9\x18\x9b\xc3E\x18\xf2\x82\xfd\xf3$\x9d_\v\nem\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00B"},
	{"132f7619d33b523b1d9e5bd8e0928355", "How can you write a big system without C++?  -Paul Glick", "md5\x01gE#\x01\xefͫ\x89\x98\xba\xdc\xfe\x102TvHow can you write a big syst\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x1c"},
}

func TestGolden(t *testing.T) {
	for i := 0; i < len(golden); i++ {
		g := golden[i]
		s := fmt.Sprintf("%x", Sum([]byte(g.in)))
		if s != g.out {
			t.Fatalf("Sum function: md5(%s) = %s want %s", g.in, s, g.out)
		}
		c := New()
		buf := make([]byte, len(g.in)+4)
		for j := 0; j < 3+4; j++ {
			if j < 2 {
				io.WriteString(c, g.in)
			} else if j == 2 {
				io.WriteString(c, g.in[:len(g.in)/2])
				c.Sum(nil)
				io.WriteString(c, g.in[len(g.in)/2:])
			} else if j > 2 {
				// test unaligned write
				buf = buf[1:]
				copy(buf, g.in)
				c.Write(buf[:len(g.in)])
			}
			s := fmt.Sprintf("%x", c.Sum(nil))
			if s != g.out {
				t.Fatalf("md5[%d](%s) = %s want %s", j, g.in, s, g.out)
			}
			c.Reset()
		}
	}
}

func TestGoldenMarshal(t *testing.T) {
	for _, g := range golden {
		h := New()
		h2 := New()

		io.WriteString(h, g.in[:len(g.in)/2])

		state, err := h.(encoding.BinaryMarshaler).MarshalBinary()
		if err != nil {
			t.Errorf("could not marshal: %v", err)
			continue
		}

		stateAppend, err := h.(encoding.BinaryAppender).AppendBinary(make([]byte, 4, 32))
		if err != nil {
			t.Errorf("could not marshal: %v", err)
			continue
		}
		stateAppend = stateAppend[4:]

		if string(state) != g.halfState {
			t.Errorf("md5(%q) state = %q, want %q", g.in, state, g.halfState)
			continue
		}

		if string(stateAppend) != g.halfState {
			t.Errorf("md5(%q) stateAppend = %q, want %q", g.in, stateAppend, g.halfState)
			continue
		}

		if err := h2.(encoding.BinaryUnmarshaler).UnmarshalBinary(state); err != nil {
			t.Errorf("could not unmarshal: %v", err)
			continue
		}

		io.WriteString(h, g.in[len(g.in)/2:])
		io.WriteString(h2, g.in[len(g.in)/2:])

		if actual, actual2 := h.Sum(nil), h2.Sum(nil); !bytes.Equal(actual, actual2) {
			t.Errorf("md5(%q) = 0x%x != marshaled 0x%x", g.in, actual, actual2)
		}
	}
}

func TestLarge(t *testing.T) {
	const N = 10000
	const offsets = 4
	ok := "2bb571599a4180e1d542f76904adc3df" // md5sum of "0123456789" * 1000
	block := make([]byte, N+offsets)
	c := New()
	for offset := 0; offset < offsets; offset++ {
		for i := 0; i < N; i++ {
			block[offset+i] = '0' + byte(i%10)
		}
		for blockSize := 10; blockSize <= N; blockSize *= 10 {
			blocks := N / blockSize
			b := block[offset : offset+blockSize]
			c.Reset()
			for i := 0; i < blocks; i++ {
				c.Write(b)
			}
			s := fmt.Sprintf("%x", c.Sum(nil))
			if s != ok {
				t.Fatalf("md5 TestLarge offset=%d, blockSize=%d = %s want %s", offset, blockSize, s, ok)
			}
		}
	}
}

func TestExtraLarge(t *testing.T) {
	const N = 100000
	const offsets = 4
	ok := "13572e9e296cff52b79c52148313c3a5" // md5sum of "0123456789" * 10000
	block := make([]byte, N+offsets)
	c := New()
	for offset := 0; offset < offsets; offset++ {
		for i := 0; i < N; i++ {
			block[offset+i] = '0' + byte(i%10)
		}
		for blockSize := 10; blockSize <= N; blockSize *= 10 {
			blocks := N / blockSize
			b := block[offset : offset+blockSize]
			c.Reset()
			for i := 0; i < blocks; i++ {
				c.Write(b)
			}
			s := fmt.Sprintf("%x", c.Sum(nil))
			if s != ok {
				t.Fatalf("md5 TestExtraLarge offset=%d, blockSize=%d = %s want %s", offset, blockSize, s, ok)
			}
		}
	}
}

// Tests that blockGeneric (pure Go) and block (in assembly for amd64, 386, arm) match.
func TestBlockGeneric(t *testing.T) {
	gen, asm := New().(*digest), New().(*digest)
	buf := make([]byte, BlockSize*20) // arbitrary factor
	rand.Read(buf)
	blockGeneric(gen, buf)
	block(asm, buf)
	if *gen != *asm {
		t.Error("block and blockGeneric resulted in different states")
	}
}

// Tests for unmarshaling hashes that have hashed a large amount of data
// The initial hash generation is omitted from the test, because it takes a long time.
// The test contains some already-generated states, and their expected sums
// Tests a problem that is outlined in GitHub issue #29541
// The problem is triggered when an amount of data has been hashed for which
// the data length has a 1 in the 32nd bit. When casted to int, this changes
// the sign of the value, and causes the modulus operation to return a
// different result.
type unmarshalTest struct {
	state string
	sum   string
}

var largeUnmarshalTests = []unmarshalTest{
	// Data length: 7_102_415_735
	{
		state: "md5\x01\xa5\xf7\xf0=\xd6S\x85\xd9M\n}\xc3\u0601\x89\xe7@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuv\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\xa7VCw",
		sum:   "cddefcf74ffec709a0b45a6a987564d5",
	},
	// Data length: 6_565_544_823
	{
		state: "md5\x01{\xda\x1a\xc7\xc9'?\x83EX\xe0\x88q\xfeG\x18@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuv\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x87VCw",
		sum:   "fd9f41874ab240698e7bc9c3ae70c8e4",
	},
}

func safeSum(h hash.Hash) (sum []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sum panic: %v", r)
		}
	}()

	return h.Sum(nil), nil
}

func TestLargeHashes(t *testing.T) {
	for i, test := range largeUnmarshalTests {

		h := New()
		if err := h.(encoding.BinaryUnmarshaler).UnmarshalBinary([]byte(test.state)); err != nil {
			t.Errorf("test %d could not unmarshal: %v", i, err)
			continue
		}

		sum, err := safeSum(h)
		if err != nil {
			t.Errorf("test %d could not sum: %v", i, err)
			continue
		}

		if fmt.Sprintf("%x", sum) != test.sum {
			t.Errorf("test %d sum mismatch: expect %s got %x", i, test.sum, sum)
		}
	}
}

func TestAllocations(t *testing.T) {
	cryptotest.SkipTestAllocations(t)
	in := []byte("hello, world!")
	out := make([]byte, 0, Size)
	h := New()
	n := int(testing.AllocsPerRun(10, func() {
		h.Reset()
		h.Write(in)
		out = h.Sum(out[:0])
	}))
	if n > 0 {
		t.Errorf("allocs = %d, want 0", n)
	}
}

func TestMD5Hash(t *testing.T) {
	cryptotest.TestHash(t, New)
}

func TestExtraMethods(t *testing.T) {
	h := maybeCloner(New())
	cryptotest.NoExtraMethods(t, &h, "MarshalBinary", "UnmarshalBinary", "AppendBinary")
}

func maybeCloner(h hash.Hash) any {
	if c, ok := h.(hash.Cloner); ok {
		return &c
	}
	return &h
}

var bench = New()
var buf = make([]byte, 1024*1024*8+1)
var sum = make([]byte, bench.Size())

func benchmarkSize(b *testing.B, size int, unaligned bool) {
	b.SetBytes(int64(size))
	buf := buf
	if unaligned {
		if uintptr(unsafe.Pointer(&buf[0]))&(unsafe.Alignof(uint32(0))-1) == 0 {
			buf = buf[1:]
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bench.Reset()
		bench.Write(buf[:size])
		bench.Sum(sum[:0])
	}
}

func BenchmarkHash8Bytes(b *testing.B) {
	benchmarkSize(b, 8, false)
}

func BenchmarkHash64(b *testing.B) {
	benchmarkSize(b, 64, false)
}

func BenchmarkHash128(b *testing.B) {
	benchmarkSize(b, 128, false)
}

func BenchmarkHash256(b *testing.B) {
	benchmarkSize(b, 256, false)
}

func BenchmarkHash512(b *testing.B) {
	benchmarkSize(b, 512, false)
}

func BenchmarkHash1K(b *testing.B) {
	benchmarkSize(b, 1024, false)
}

func BenchmarkHash8K(b *testing.B) {
	benchmarkSize(b, 8192, false)
}

func BenchmarkHash1M(b *testing.B) {
	benchmarkSize(b, 1024*1024, false)
}

func BenchmarkHash8M(b *testing.B) {
	benchmarkSize(b, 8*1024*1024, false)
}

func BenchmarkHash8BytesUnaligned(b *testing.B) {
	benchmarkSize(b, 8, true)
}

func BenchmarkHash1KUnaligned(b *testing.B) {
	benchmarkSize(b, 1024, true)
}

func BenchmarkHash8KUnaligned(b *testing.B) {
	benchmarkSize(b, 8192, true)
}
