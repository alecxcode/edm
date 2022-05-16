package main

import (
	"database/sql"
	"math/rand"
	"strconv"
	"time"
)

func fillDBwithTestData(db *sql.DB, DBType byte) {

	rand.Seed(time.Now().UnixNano())

	company := Company{
		ShortName:   "Example Company",
		FullName:    "Example Company Corporation",
		ForeignName: "Example Company Corp.",
		Contacts: CompanyContacts{
			AddressReg:  "US, NY, 11 some avenue",
			AddressFact: "US, NY, 11 some avenue",
			Phone:       "+1(111)111 02 03",
			Email:       "info@example.com",
			WebSite:     "https://www.example.com",
			Other:       "",
		},
		RegNo:       "",
		TaxNo:       "",
		BankDetails: "",
	}
	cid, _ := company.create(db, DBType)
	unitA := Unit{UnitName: "Archives and Documentation Office", Company: &Company{ID: cid}}
	archunit, _ := unitA.create(db, DBType)
	unitB := Unit{UnitName: "Engineering Team", Company: &Company{ID: cid}}
	unitB.create(db, DBType)
	unitC := Unit{UnitName: "Research and Development Unit", Company: &Company{ID: cid}}
	unitC.create(db, DBType)
	unitD := Unit{UnitName: "Software Developers", Company: &Company{ID: cid}}
	softunit, _ := unitD.create(db, DBType)
	unitE := Unit{UnitName: "West Branch Office", Company: &Company{ID: cid}}
	unitE.create(db, DBType)

	defaultUserConfig := UserConfig{
		SystemTheme:          "dark",
		ElemsOnPage:          20,
		ElemsOnPageTeam:      500,
		DateFormat:           "dd.mm.yyyy",
		TimeFormat:           "24h",
		UseCalendarInConrols: true,
		CurrencyBeforeAmount: true,
	}
	user1 := Profile{
		FirstName: "John",
		OtherName: "",
		Surname:   "Smith",
		Contacts: UserContacts{
			TelOffice: "333",
			TelMobile: "",
			Email:     "",
		},
		BirthDate:  Date{1990, 10, 22},
		JobTitle:   "Software developer",
		JobUnit:    &Unit{ID: softunit},
		UserConfig: defaultUserConfig,
	}
	user1.create(db, DBType)
	user2 := Profile{
		FirstName: "Jane",
		OtherName: "",
		Surname:   "Anderson",
		Contacts: UserContacts{
			TelOffice: "442",
			TelMobile: "",
			Email:     "",
		},
		BirthDate:  Date{1900, 12, 25},
		JobTitle:   "Document archivist",
		JobUnit:    &Unit{ID: archunit},
		UserConfig: defaultUserConfig,
	}
	user2.create(db, DBType)

	for j := 0; j < 2; j++ {
		cat := 0
		for i := 0; i < 27; i++ {
			doctype := rand.Intn(13)
			month := rand.Intn(11) + 1
			cat++
			if cat > 3 {
				cat = 0
			}
			d := Document{
				RegNo:     "rno-" + strconv.Itoa(i+1),
				RegDate:   Date{1990 + i, byte(month), byte(i + 1)},
				IncNo:     "ino-" + strconv.Itoa(i+10),
				IncDate:   Date{1990 + i, byte(month), byte(i + 2)},
				Category:  cat,
				DocType:   doctype,
				About:     "Test content about. Some test text. Document description or topic.",
				Authors:   "Some authors who made or signed the document.",
				Addressee: "Some text to whom it is addressed.",
				DocSum:    100 * i,
				Currency:  840,
				EndDate:   Date{2000 + i, byte(month), byte(i)},
				Creator:   &Profile{ID: 2},
				Note:      "",
			}
			d.create(db, DBType)
		}
	}

	for j := 0; j < 2; j++ {
		tasktopics := map[int]string{
			0: "Check some test task",
			1: "Do a job with some task",
			2: "Test the app",
			3: "Make some changes",
			4: "Test app function",
			5: "See how BB-code works",
			6: "Verify the document",
			7: "Do some calculations",
			8: "Edit some object",
			9: "Other task to do",
		}
		taskcontent := map[int]string{
			0: "This is test task and test task content.",
			1: "Do a test task. Some test text.",
			2: "Check if the program works.",
			3: "There is a situation here:\n [b]We need to fix the bugs.[/b]",
			4: "Check how files are uploaded.",
			5: "This is an example of [b]bold font[/b], or [u]underline font[/u], or a preformatted chunk of text: [code]\nif (a == b){\n  console.log('hello, world!')\n}[/code]\nThat's how it works.",
			6: "View and check the document.",
			7: "Make some mathematical or financial calculations.",
			8: "Test task content.",
			9: "To do something else.",
		}
		for i := 0; i < 27; i++ {
			month := 1
			if i > 10 {
				month = 2
			}
			if i > 20 {
				month = 3
			}
			day := i / 2
			if day < 1 {
				day = 1
			}
			if day > 28 {
				day = 28
			}
			v := rand.Intn(9)
			s := rand.Intn(5) + 1
			hour := i + 1
			if hour > 23 {
				hour = 23
			}
			minute := rand.Intn(59) + 1
			minutex := rand.Intn(59) + 1
			creator := rand.Intn(3) + 1
			assignee := rand.Intn(3) + 1
			var tid int
			if creator == 1 {
				t := Task{
					Created:    DateTime{2022, 1, byte(day), byte(hour), byte(i + 3)},
					PlanStart:  DateTime{2022, byte(month + 1), byte(day), byte(hour), byte(minute)},
					PlanDue:    DateTime{2022, byte(month + 2), byte(day), byte(hour), byte(minute)},
					StatusSet:  DateTime{2022, 1, byte(day), byte(hour), byte(minutex)},
					Creator:    &Profile{ID: creator},
					Assignee:   nil,
					Topic:      tasktopics[v],
					Content:    taskcontent[v],
					TaskStatus: 0,
				}
				tid, _ = t.create(db, DBType)
			} else {
				t := Task{
					Created:    DateTime{2022, 1, byte(day), byte(hour), byte(i + 3)},
					PlanStart:  DateTime{2022, byte(month + 1), byte(day), byte(hour), byte(minute)},
					PlanDue:    DateTime{2022, byte(month + 2), byte(day), byte(hour), byte(minute)},
					StatusSet:  DateTime{2022, 1, byte(day), byte(hour), byte(minutex)},
					Creator:    &Profile{ID: creator},
					Assignee:   &Profile{ID: assignee},
					Topic:      tasktopics[v],
					Content:    taskcontent[v],
					TaskStatus: s,
				}
				tid, _ = t.create(db, DBType)
			}
			for x := 0; x < 5; x++ {
				m := x + 1
				if m > 12 {
					m = 12
				}
				comment := Comment{
					Task:    &Task{ID: tid},
					Creator: &Profile{ID: creator},
					Created: DateTime{2022, byte(m), byte(x + 1), byte(x + 2), byte(i + 5)},
					Content: "Test comment iteration: 000" + strconv.Itoa(i) + strconv.Itoa(x),
				}
				comment.create(db, DBType)
			}
		}
	}
}
