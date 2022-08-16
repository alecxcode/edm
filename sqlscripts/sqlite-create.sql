CREATE TABLE IF NOT EXISTS companies
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
ShortName TEXT,
FullName TEXT,
ForeignName TEXT,
Contacts TEXT,
CompanyHead INTEGER,
RegNo TEXT,
TaxNo TEXT,
BankDetails TEXT,
CONSTRAINT fk_companies_CompanyHead FOREIGN KEY (CompanyHead) REFERENCES profiles(ID) ON DELETE SET NULL);

CREATE TABLE IF NOT EXISTS units
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
UnitName TEXT,
Company INTEGER,
UnitHead INTEGER,
CONSTRAINT fk_units_Company FOREIGN KEY (Company) REFERENCES companies(ID) ON DELETE CASCADE,
CONSTRAINT fk_units_UnitHead FOREIGN KEY (UnitHead) REFERENCES profiles(ID) ON DELETE SET NULL);

CREATE TABLE IF NOT EXISTS documents
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
RegNo TEXT,
RegDate INTEGER,
IncNo TEXT,
IncDate INTEGER,
Category INTEGER,
DocType INTEGER,
About TEXT,
Authors TEXT,
Addressee TEXT,
DocSum INTEGER,
Currency INTEGER,
EndDate INTEGER,
Creator INTEGER,
Note TEXT,
FileList TEXT,
CONSTRAINT fk_documents_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL);

CREATE INDEX idx_documents_RegDate ON documents (RegDate);

CREATE INDEX idx_documents_IncDate ON documents (IncDate);

CREATE TABLE IF NOT EXISTS approvals
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
Written INTEGER,
Approver INTEGER,
ApproverSign TEXT,
DocID INTEGER,
Approved INTEGER,
Note TEXT,
CONSTRAINT fk_approvals_Approver FOREIGN KEY (Approver) REFERENCES profiles(ID) ON DELETE SET NULL,
CONSTRAINT fk_approvals_DocID FOREIGN KEY (DocID) REFERENCES documents(ID) ON DELETE CASCADE);

CREATE INDEX idx_approvals_DocID ON approvals (DocID);

CREATE TABLE IF NOT EXISTS profiles
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
FirstName TEXT,
OtherName TEXT,
Surname TEXT,
BirthDate INTEGER,
JobTitle TEXT,
JobUnit INTEGER,
Boss INTEGER,
Contacts TEXT,
UserRole INTEGER,
UserLock INTEGER,
UserConfig TEXT,
Login TEXT,
Passwd TEXT,
CONSTRAINT fk_profiles_JobUnit FOREIGN KEY (JobUnit) REFERENCES units(ID) ON DELETE SET NULL,
CONSTRAINT fk_profiles_Boss FOREIGN KEY (Boss) REFERENCES profiles(ID) ON DELETE SET NULL);

CREATE TABLE IF NOT EXISTS emailmessages
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
SendTo TEXT,
SendCc TEXT,
Subj TEXT,
Cont TEXT);

CREATE TABLE IF NOT EXISTS tasks
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
Created INTEGER,
PlanStart INTEGER,
PlanDue INTEGER,
StatusSet INTEGER,
Creator INTEGER,
Assignee INTEGER,
Participants TEXT,
Topic TEXT,
Content TEXT,
TaskStatus INTEGER,
Project INTEGER,
FileList TEXT,
CONSTRAINT fk_tasks_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL,
CONSTRAINT fk_tasks_Assignee FOREIGN KEY (Assignee) REFERENCES profiles(ID) ON DELETE SET NULL);

CREATE INDEX idx_tasks_Created ON tasks (Created);

CREATE INDEX idx_tasks_PlanStart ON tasks (PlanStart);

CREATE INDEX idx_tasks_PlanDue ON tasks (PlanDue);

CREATE INDEX idx_tasks_StatusSet ON tasks (StatusSet);

CREATE TABLE IF NOT EXISTS comments
(ID INTEGER PRIMARY KEY AUTOINCREMENT,
Created INTEGER,
Creator INTEGER,
Task INTEGER,
Content TEXT,
FileList TEXT,
CONSTRAINT fk_comments_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL,
CONSTRAINT fk_comments_Task FOREIGN KEY (Task) REFERENCES tasks(ID) ON DELETE CASCADE);

CREATE INDEX idx_comments_Task ON comments (Task);

