CREATE TABLE  approvals
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
Written INTEGER,
Approver INTEGER,
ApproverSign VARCHAR2(2000),
DocID INTEGER,
Approved INTEGER,
Note CLOB);

CREATE INDEX idx_approvals_DocID ON approvals (DocID);

CREATE TABLE  documents
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
RegNo VARCHAR2(255),
RegDate INTEGER,
IncNo VARCHAR2(255),
IncDate INTEGER,
Category INTEGER,
DocType INTEGER,
About VARCHAR2(4000),
Authors VARCHAR2(2000),
Addressee VARCHAR2(2000),
DocSum INTEGER,
Currency INTEGER,
EndDate INTEGER,
Creator INTEGER,
Note VARCHAR2(4000),
FileList CLOB);

CREATE INDEX idx_documents_RegDate ON documents (RegDate);

CREATE INDEX idx_documents_IncDate ON documents (IncDate);

CREATE TABLE  emailmessages
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
SendTo CLOB,
SendCc CLOB,
Subj VARCHAR2(4000),
Cont CLOB);

CREATE TABLE  projects
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
ProjName VARCHAR2(255),
Description CLOB,
Creator INTEGER,
ProjStatus INTEGER);

CREATE TABLE  comments
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
Created INTEGER,
Creator INTEGER,
Task INTEGER,
Content CLOB,
FileList CLOB);

CREATE INDEX idx_comments_Task ON comments (Task);

CREATE TABLE  tasks
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
Created INTEGER,
PlanStart INTEGER,
PlanDue INTEGER,
StatusSet INTEGER,
Creator INTEGER,
Assignee INTEGER,
Participants VARCHAR2(4000),
Topic VARCHAR2(255),
Content CLOB,
TaskStatus INTEGER,
Project INTEGER,
FileList CLOB);

CREATE INDEX idx_tasks_Created ON tasks (Created);

CREATE INDEX idx_tasks_PlanStart ON tasks (PlanStart);

CREATE INDEX idx_tasks_PlanDue ON tasks (PlanDue);

CREATE INDEX idx_tasks_StatusSet ON tasks (StatusSet);

CREATE INDEX idx_tasks_Project ON tasks (Project);

CREATE TABLE  companies
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
ShortName VARCHAR2(255),
FullName VARCHAR2(512),
ForeignName VARCHAR2(512),
Contacts VARCHAR2(4000),
CompanyHead INTEGER,
RegNo VARCHAR2(255),
TaxNo VARCHAR2(255),
BankDetails VARCHAR2(4000));

CREATE TABLE  profiles
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
FirstName VARCHAR2(255),
OtherName VARCHAR2(255),
Surname VARCHAR2(255),
BirthDate INTEGER,
JobTitle VARCHAR2(255),
JobUnit INTEGER,
Boss INTEGER,
Contacts VARCHAR2(4000),
UserRole INTEGER,
UserLock INTEGER,
UserConfig VARCHAR2(4000),
Login VARCHAR2(255),
Passwd VARCHAR2(255));

CREATE TABLE  units
(ID INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
UnitName VARCHAR2(1024),
Company INTEGER,
UnitHead INTEGER);

ALTER TABLE approvals ADD CONSTRAINT fk_approvals_Approver FOREIGN KEY (Approver) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE approvals ADD CONSTRAINT fk_approvals_DocID FOREIGN KEY (DocID) REFERENCES documents(ID) ON DELETE CASCADE;
ALTER TABLE documents ADD CONSTRAINT fk_documents_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE projects ADD CONSTRAINT fk_projects_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE comments ADD CONSTRAINT fk_comments_Task FOREIGN KEY (Task) REFERENCES tasks(ID) ON DELETE CASCADE;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_Creator FOREIGN KEY (Creator) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_Assignee FOREIGN KEY (Assignee) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_Project FOREIGN KEY (Project) REFERENCES projects(ID) ON DELETE SET NULL;
ALTER TABLE companies ADD CONSTRAINT fk_companies_CompanyHead FOREIGN KEY (CompanyHead) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE profiles ADD CONSTRAINT fk_profiles_JobUnit FOREIGN KEY (JobUnit) REFERENCES units(ID) ON DELETE SET NULL;
ALTER TABLE profiles ADD CONSTRAINT fk_profiles_Boss FOREIGN KEY (Boss) REFERENCES profiles(ID) ON DELETE SET NULL;
ALTER TABLE units ADD CONSTRAINT fk_units_Company FOREIGN KEY (Company) REFERENCES companies(ID) ON DELETE CASCADE;
ALTER TABLE units ADD CONSTRAINT fk_units_UnitHead FOREIGN KEY (UnitHead) REFERENCES profiles(ID) ON DELETE SET NULL;
