package model

import "testing"

func TestHistory_Update(t *testing.T) {
	tests := []struct {
		name           string
		initialHistory History
		from           string
		to             string
		by             string
		comment        string
		otherHistory   History
		wantUpdate     bool
		wantHistory    []HistoryRecord
	}{
		{
			name:           "Status change creates history",
			initialHistory: History{},
			from:           "A",
			to:             "B",
			by:             "test",
			comment:        "test comment",
			otherHistory:   History{},
			wantUpdate:     true,
			wantHistory: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test",
					Comment: "test comment",
					Updated: Now(),
				},
			},
		},
		{
			name:           "Comment only creates history",
			initialHistory: History{},
			from:           "A",
			to:             "",
			by:             "test",
			comment:        "test comment",
			otherHistory:   History{},
			wantUpdate:     false,
			wantHistory: []HistoryRecord{
				{
					By:      "test",
					Comment: "test comment",
					Updated: Now(),
				},
			},
		},
		{
			name: "Remove comment-only history entry",
			initialHistory: History{
				History: []HistoryRecord{
					{
						By:      "test1",
						Comment: "first comment",
						Updated: Now(),
					},
					{
						By:      "test2",
						Comment: "second comment",
						Updated: Now(),
					},
				},
			},
			otherHistory: History{
				Remove: &[]int{0}[0],
			},
			wantUpdate: false,
			wantHistory: []HistoryRecord{
				{
					By:      "test2",
					Comment: "second comment",
					Updated: Now(),
				},
			},
		},
		{
			name: "Clear comment on status change entry",
			initialHistory: History{
				History: []HistoryRecord{
					{
						From:    "A",
						To:      "B",
						By:      "test1",
						Comment: "first change",
						Updated: Now(),
					},
					{
						By:      "test2",
						Comment: "just a comment",
						Updated: Now(),
					},
				},
			},
			otherHistory: History{
				Remove: &[]int{0}[0],
			},
			wantUpdate: false,
			wantHistory: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test1",
					Comment: "",
					Updated: Now(),
				},
				{
					By:      "test2",
					Comment: "just a comment",
					Updated: Now(),
				},
			},
		},
		{
			name: "Remove middle comment-only entry",
			initialHistory: History{
				History: []HistoryRecord{
					{
						From:    "A",
						To:      "B",
						By:      "test1",
						Comment: "status change",
						Updated: Now(),
					},
					{
						By:      "test2",
						Comment: "middle comment",
						Updated: Now(),
					},
					{
						From:    "B",
						To:      "C",
						By:      "test3",
						Comment: "another change",
						Updated: Now(),
					},
				},
			},
			otherHistory: History{
				Remove: &[]int{1}[0],
			},
			wantUpdate: false,
			wantHistory: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test1",
					Comment: "status change",
					Updated: Now(),
				},
				{
					From:    "B",
					To:      "C",
					By:      "test3",
					Comment: "another change",
					Updated: Now(),
				},
			},
		},
		{
			name: "Remove last comment-only entry",
			initialHistory: History{
				History: []HistoryRecord{
					{
						From:    "A",
						To:      "B",
						By:      "test1",
						Comment: "status change",
						Updated: Now(),
					},
					{
						By:      "test2",
						Comment: "final comment",
						Updated: Now(),
					},
				},
			},
			otherHistory: History{
				Remove: &[]int{1}[0],
			},
			wantUpdate: false,
			wantHistory: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test1",
					Comment: "status change",
					Updated: Now(),
				},
			},
		},
		{
			name: "Invalid removal index does nothing",
			initialHistory: History{
				History: []HistoryRecord{
					{
						From:    "A",
						To:      "B",
						By:      "test1",
						Comment: "status change",
						Updated: Now(),
					},
				},
			},
			otherHistory: History{
				Remove: &[]int{99}[0],
			},
			wantUpdate: false,
			wantHistory: []HistoryRecord{
				{
					From:    "A",
					To:      "B",
					By:      "test1",
					Comment: "status change",
					Updated: Now(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize empty history array if nil
			if tt.initialHistory.History == nil {
				tt.initialHistory.History = []HistoryRecord{}
			}

			gotUpdate := tt.initialHistory.Update(tt.from, tt.to, tt.by, tt.comment, tt.otherHistory)
			if gotUpdate != tt.wantUpdate {
				t.Errorf("History.Update() update = %v, want %v", gotUpdate, tt.wantUpdate)
			}

			if len(tt.initialHistory.History) != len(tt.wantHistory) {
				t.Errorf("History.Update() history length = %v, want %v", len(tt.initialHistory.History), len(tt.wantHistory))
				t.Errorf("Got history: %+v", tt.initialHistory.History)
				t.Errorf("Want history: %+v", tt.wantHistory)
				return
			}

			for i := range tt.wantHistory {
				if tt.initialHistory.History[i].From != tt.wantHistory[i].From {
					t.Errorf("History[%d].From = %v, want %v", i, tt.initialHistory.History[i].From, tt.wantHistory[i].From)
				}
				if tt.initialHistory.History[i].To != tt.wantHistory[i].To {
					t.Errorf("History[%d].To = %v, want %v", i, tt.initialHistory.History[i].To, tt.wantHistory[i].To)
				}
				if tt.initialHistory.History[i].Comment != tt.wantHistory[i].Comment {
					t.Errorf("History[%d].Comment = %v, want %v", i, tt.initialHistory.History[i].Comment, tt.wantHistory[i].Comment)
				}
				if tt.initialHistory.History[i].By != tt.wantHistory[i].By {
					t.Errorf("History[%d].By = %v, want %v", i, tt.initialHistory.History[i].By, tt.wantHistory[i].By)
				}
			}
		})
	}
}
