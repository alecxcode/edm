package filldata

import (
	"database/sql"
	"edm/internal/docs"
	"edm/internal/projs"
	"edm/internal/tasks"
	"edm/internal/team"
	"edm/pkg/datetime"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// FillDBwithTestData is for showcase or testing
func FillDBwithTestData(db *sql.DB, DBType byte) {

	log.Println("Populating DB with test data...")

	rand.Seed(time.Now().UnixNano())

	company := team.Company{
		ShortName:   "Example Company",
		FullName:    "Example Company Corporation",
		ForeignName: "Example Company Corp.",
		Contacts: team.CompanyContacts{
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
	cid, _ := company.Create(db, DBType)
	unitA := team.Unit{UnitName: "Archives and Documentation Office", Company: &team.Company{ID: cid}}
	archunit, _ := unitA.Create(db, DBType)
	unitB := team.Unit{UnitName: "Engineering Team", Company: &team.Company{ID: cid}}
	unitB.Create(db, DBType)
	unitC := team.Unit{UnitName: "Research and Development Unit", Company: &team.Company{ID: cid}}
	unitC.Create(db, DBType)
	unitD := team.Unit{UnitName: "Software Developers", Company: &team.Company{ID: cid}}
	softunit, _ := unitD.Create(db, DBType)
	unitE := team.Unit{UnitName: "West Branch Office", Company: &team.Company{ID: cid}}
	unitE.Create(db, DBType)

	defaultUserConfig := team.UserConfig{
		SystemTheme:           "dark",
		ElemsOnPage:           20,
		ElemsOnPageTeam:       500,
		DateFormat:            "dd.mm.yyyy",
		TimeFormat:            "24h",
		UseCalendarInControls: true,
		CurrencyBeforeAmount:  true,
	}
	user1 := team.Profile{
		FirstName: "John",
		OtherName: "",
		Surname:   "Smith",
		Contacts: team.UserContacts{
			TelOffice: "333",
			TelMobile: "",
			Email:     "",
		},
		BirthDate:  datetime.Date{Year: 1990, Month: 10, Day: 22},
		JobTitle:   "Software developer",
		JobUnit:    &team.Unit{ID: softunit},
		UserConfig: defaultUserConfig,
	}
	uid1, _ := user1.Create(db, DBType)
	user2 := team.Profile{
		FirstName: "Jane",
		OtherName: "",
		Surname:   "Anderson",
		Contacts: team.UserContacts{
			TelOffice: "442",
			TelMobile: "",
			Email:     "",
		},
		BirthDate:  datetime.Date{Year: 1900, Month: 12, Day: 25},
		JobTitle:   "Document archivist",
		JobUnit:    &team.Unit{ID: archunit},
		UserConfig: defaultUserConfig,
	}
	uid2, _ := user2.Create(db, DBType)

	for j := 0; j < 2; j++ {
		cat := 0
		for i := 0; i < 27; i++ {
			doctype := rand.Intn(13)
			month := rand.Intn(11) + 1
			cat++
			if cat > 3 {
				cat = 0
			}
			d := docs.Document{
				RegNo:     "rno-" + strconv.Itoa(i+1),
				RegDate:   datetime.Date{Year: 1990 + i, Month: byte(month), Day: byte(i + 1)},
				IncNo:     "ino-" + strconv.Itoa(i+10),
				IncDate:   datetime.Date{Year: 1990 + i, Month: byte(month), Day: byte(i + 2)},
				Category:  cat,
				DocType:   doctype,
				About:     "Test content about. Some test text. Document description or topic.",
				Authors:   "Some authors who made or signed the document.",
				Addressee: "Some text to whom it is addressed.",
				DocSum:    100 * i,
				Currency:  840,
				EndDate:   datetime.Date{Year: 2000 + i, Month: byte(month), Day: byte(i)},
				Creator:   &team.Profile{ID: uid2},
				Note:      "",
			}
			d.Create(db, DBType)
		}
	}

	project1 := projs.Project{
		ProjName:    "Some test project",
		Description: "Description of the test project",
		Creator:     &team.Profile{ID: uid1},
		ProjStatus:  0,
	}
	projid1, _ := project1.Create(db, DBType)
	project2 := projs.Project{
		ProjName:    "Another test project",
		Description: "Some text to describe the project",
		Creator:     &team.Profile{ID: uid2},
		ProjStatus:  0,
	}
	projid2, _ := project2.Create(db, DBType)

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
			project := 0
			if i < 11 && j == 0 {
				project = projid1
			}
			if i < 11 && j == 1 {
				project = projid2
			}
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
			s := rand.Intn(6) + 1
			hour := i + 1
			if hour > 23 {
				hour = 23
			}
			minute := rand.Intn(59) + 1
			minutex := rand.Intn(59) + 1

			creator := 0
			assignee := 0
			creatorRand := rand.Intn(3) + 1
			switch creatorRand {
			case 1:
				creator = 1
			case 2:
				creator = uid1
			case 3:
				creator = uid2
			}
			assigneeRand := rand.Intn(3) + 1
			switch assigneeRand {
			case 1:
				assignee = 1
			case 2:
				assignee = uid1
			case 3:
				assignee = uid2
			}

			var tid int
			if creator == 1 {
				t := tasks.Task{
					Created:    datetime.DateTime{Year: 2022, Month: 1, Day: byte(day), Hour: byte(hour), Minute: byte(i + 3)},
					PlanStart:  datetime.DateTime{Year: 2022, Month: byte(month + 1), Day: byte(day), Hour: byte(hour), Minute: byte(minute)},
					PlanDue:    datetime.DateTime{Year: 2022, Month: byte(month + 2), Day: byte(day), Hour: byte(hour), Minute: byte(minute)},
					StatusSet:  datetime.DateTime{Year: 2022, Month: 1, Day: byte(day), Hour: byte(hour), Minute: byte(minutex)},
					Creator:    &team.Profile{ID: creator},
					Assignee:   nil,
					Topic:      tasktopics[v],
					Content:    taskcontent[v],
					TaskStatus: 0,
					Project:    project,
				}
				tid, _ = t.Create(db, DBType)
			} else {
				t := tasks.Task{
					Created:    datetime.DateTime{Year: 2022, Month: 1, Day: byte(day), Hour: byte(hour), Minute: byte(i + 3)},
					PlanStart:  datetime.DateTime{Year: 2022, Month: byte(month + 1), Day: byte(day), Hour: byte(hour), Minute: byte(minute)},
					PlanDue:    datetime.DateTime{Year: 2022, Month: byte(month + 2), Day: byte(day), Hour: byte(hour), Minute: byte(minute)},
					StatusSet:  datetime.DateTime{Year: 2022, Month: 1, Day: byte(day), Hour: byte(hour), Minute: byte(minutex)},
					Creator:    &team.Profile{ID: creator},
					Assignee:   &team.Profile{ID: assignee},
					Topic:      tasktopics[v],
					Content:    taskcontent[v],
					TaskStatus: s,
					Project:    project,
				}
				tid, _ = t.Create(db, DBType)
			}
			for x := 0; x < 5; x++ {
				m := x + 1
				if m > 12 {
					m = 12
				}
				comment := tasks.Comment{
					Task:    tid,
					Creator: &team.Profile{ID: creator},
					Created: datetime.DateTime{Year: 2022, Month: byte(m), Day: byte(x + 1), Hour: byte(x + 2), Minute: byte(i + 5)},
					Content: "Test comment iteration: 000" + strconv.Itoa(i) + strconv.Itoa(x),
				}
				comment.Create(db, DBType)
			}
		}
	}
}
