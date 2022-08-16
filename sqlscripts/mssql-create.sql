CREATE TABLE  companies
(ID INTEGER IDENTITY PRIMARY KEY,
ShortName VARCHAR(255),
FullName VARCHAR(512),
ForeignName VARCHAR(512),
Contacts VARCHAR(4000),
CompanyHead INTEGER,
RegNo VARCHAR(255),
TaxNo VARCHAR(255),
BankDetails VARCHAR(4000));

CREATE TABLE  units
(ID INTEGER IDENTITY PRIMARY KEY,
UnitName VARCHAR(1024),
Company INTEGER,
UnitHead INTEGER);

CREATE TABLE  documents
(ID INTEGER IDENTITY PRIMARY KEY,
RegNo VARCHAR(255),
RegDate BIGINT,
IncNo VARCHAR(255),
IncDate BIGINT,
Category INTEGER,
DocType INTEGER,
About VARCHAR(4000),
Authors VARCHAR(2000),
Addressee VARCHAR(2000),
DocSum BIGINT,
Currency INTEGER,
EndDate BIGINT,
Creator INTEGER,
Note VARCHAR(4000),
FileList VARCHAR(max));

CREATE INDEX idx_documents_RegDate ON documents (RegDate);

CREATE INDEX idx_documents_IncDate ON documents (IncDate);

CREATE TABLE  approvals
(ID INTEGER IDENTITY PRIMARY KEY,
Written BIGINT,
Approver INTEGER,
ApproverSign VARCHAR(4000),
DocID INTEGER,
Approved INTEGER,
Note VARCHAR(max));

CREATE INDEX idx_approvals_DocID ON approvals (DocID);

CREATE TABLE  profiles
(ID INTEGER IDENTITY PRIMARY KEY,
FirstName VARCHAR(255),
OtherName VARCHAR(255),
Surname VARCHAR(255),
BirthDate BIGINT,
JobTitle VARCHAR(255),
JobUnit INTEGER,
Boss INTEGER,
Contacts VARCHAR(4000),
UserRole INTEGER,
UserLock INTEGER,
UserConfig VARCHAR(4000),
Login VARCHAR(255),
Passwd VARCHAR(255));

CREATE TABLE  emailmessages
(ID INTEGER IDENTITY PRIMARY KEY,
SendTo VARCHAR(max),
SendCc VARCHAR(max),
Subj VARCHAR(4000),
Cont VARCHAR(max));

CREATE TABLE  tasks
(ID INTEGER IDENTITY PRIMARY KEY,
Created BIGINT,
PlanStart BIGINT,
PlanDue BIGINT,
StatusSet BIGINT,
Creator INTEGER,
Assignee INTEGER,
Participants VARCHAR(4000),
Topic VARCHAR(255),
Content VARCHAR(max),
TaskStatus INTEGER,
Project INTEGER,
FileList VARCHAR(max));

CREATE INDEX idx_tasks_Created ON tasks (Created);

CREATE INDEX idx_tasks_PlanStart ON tasks (PlanStart);

CREATE INDEX idx_tasks_PlanDue ON tasks (PlanDue);

CREATE INDEX idx_tasks_StatusSet ON tasks (StatusSet);

CREATE TABLE  comments
(ID INTEGER IDENTITY PRIMARY KEY,
Created BIGINT,
Creator INTEGER,
Task INTEGER,
Content VARCHAR(max),
FileList VARCHAR(max));

CREATE INDEX idx_comments_Task ON comments (Task);

ALTER TABLE companies ADD CONSTRAINT fk_companies_CompanyHead FOREIGN KEY (CompanyHead) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE units ADD CONSTRAINT fk_units_Company FOREIGN KEY (Company) REFERENCES companies(ID) ON DELETE CASCADE;
ALTER TABLE units ADD CONSTRAINT fk_units_UnitHead FOREIGN KEY (UnitHead) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE documents ADD CONSTRAINT fk_documents_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE approvals ADD CONSTRAINT fk_approvals_Approver FOREIGN KEY (Approver) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE approvals ADD CONSTRAINT fk_approvals_DocID FOREIGN KEY (DocID) REFERENCES documents(ID) ON DELETE CASCADE;
ALTER TABLE profiles ADD CONSTRAINT fk_profiles_JobUnit FOREIGN KEY (JobUnit) REFERENCES units(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Task FOREIGN KEY (Task) REFERENCES tasks(ID) ON DELETE CASCADE;
