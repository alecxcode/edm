CREATE TABLE companies
(ID SERIAL PRIMARY KEY,
ShortName TEXT,
FullName TEXT,
ForeignName TEXT,
Contacts TEXT,
CompanyHead INTEGER,
RegNo TEXT,
TaxNo TEXT,
BankDetails TEXT);

CREATE TABLE units
(ID SERIAL PRIMARY KEY,
UnitName TEXT,
Company INTEGER,
UnitHead INTEGER);

CREATE TABLE documents
(ID SERIAL PRIMARY KEY,
RegNo TEXT,
RegDate BIGINT,
IncNo TEXT,
IncDate BIGINT,
Category INTEGER,
DocType INTEGER,
About TEXT,
Authors TEXT,
Addressee TEXT,
DocSum BIGINT,
Currency INTEGER,
EndDate BIGINT,
Creator INTEGER,
Note TEXT,
FileList TEXT);

CREATE INDEX idx_documents_RegDate ON documents (RegDate);

CREATE INDEX idx_documents_IncDate ON documents (IncDate);

CREATE TABLE profiles
(ID SERIAL PRIMARY KEY,
FirstName TEXT,
OtherName TEXT,
Surname TEXT,
BirthDate BIGINT,
JobTitle TEXT,
JobUnit INTEGER,
Boss INTEGER,
Contacts TEXT,
UserRole INTEGER,
UserLock INTEGER,
UserConfig TEXT,
Login TEXT,
Passwd TEXT);

CREATE TABLE emailmessages
(ID SERIAL PRIMARY KEY,
SendTo TEXT,
SendCc TEXT,
Subj TEXT,
Cont TEXT);

CREATE TABLE tasks
(ID SERIAL PRIMARY KEY,
Created BIGINT,
PlanStart BIGINT,
PlanDue BIGINT,
StatusSet BIGINT,
Creator INTEGER,
Assignee INTEGER,
Participants TEXT,
Topic TEXT,
Content TEXT,
TaskStatus INTEGER,
Project INTEGER,
FileList TEXT);

CREATE INDEX idx_tasks_Created ON tasks (Created);

CREATE INDEX idx_tasks_PlanStart ON tasks (PlanStart);

CREATE INDEX idx_tasks_PlanDue ON tasks (PlanDue);

CREATE INDEX idx_tasks_StatusSet ON tasks (StatusSet);

CREATE TABLE comments
(ID SERIAL PRIMARY KEY,
Created BIGINT,
Creator INTEGER,
Task INTEGER,
Content TEXT,
FileList TEXT);

CREATE INDEX idx_comments_Task ON comments (Task);

ALTER TABLE companies ADD CONSTRAINT fk_companies_CompanyHead FOREIGN KEY (CompanyHead) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE units ADD CONSTRAINT fk_units_Company FOREIGN KEY (Company) REFERENCES companies(ID) ON DELETE CASCADE;
ALTER TABLE units ADD CONSTRAINT fk_units_UnitHead FOREIGN KEY (UnitHead) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE documents ADD CONSTRAINT fk_documents_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE profiles ADD CONSTRAINT fk_profiles_JobUnit FOREIGN KEY (JobUnit) REFERENCES units(ID) ON DELETE SET NULL;
ALTER TABLE profiles ADD CONSTRAINT fk_profiles_Boss FOREIGN KEY (Boss) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_Assignee FOREIGN KEY (Assignee) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Task FOREIGN KEY (Task) REFERENCES tasks(ID) ON DELETE CASCADE;
