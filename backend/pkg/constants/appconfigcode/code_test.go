package appconfigcode

import "testing"

func TestWholesaleEmailValidate(t *testing.T) {
	cases := map[string]bool{
		`{"replyTo": "owner@example.com", "cc": ["owner@example.com"]}`: true,
		`{"replyTo": "", "cc": []}`:                                     true, // no headers
		`{"replyTo": "not-an-email", "cc": []}`:                         false,
		`{"replyTo": "", "cc": ["a@example.com", "bad"]}`:               false,
		`{"replyTo": "", "cc": [""]}`:                                   false,
		`{"replyTo": "", "cc": [], "bcc": []}`:                          false, // unknown field
		`{"cc": ["1@x.co","2@x.co","3@x.co","4@x.co","5@x.co","6@x.co","7@x.co","8@x.co","9@x.co","10@x.co","11@x.co"]}`: false,
	}
	for data, ok := range cases {
		if err := AppConfigWholesaleEmail.Validate(data); (err == nil) != ok {
			t.Errorf("Validate(%s) = %v, want ok=%v", data, err, ok)
		}
	}
}
