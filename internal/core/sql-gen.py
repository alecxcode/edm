#!/usr/bin/python3

# This generates SQL script for go structs with
# the comment 'sql generate' after struct name.
# Some additional instructions for this generator
# may be provided in struct tags (see this file
# or project go source files for examples).
# Run go generate in project folder.

import os
import re

def recursiveFileList(d):
    flist = [os.path.join(d, f) for f in os.listdir(d)]
    for f in flist:
        if os.path.isdir(f):
            flist += recursiveFileList(f)
    return flist

def makeFileList():
    flist = [f for f in os.listdir() if f.endswith('.go')]
    flist += [f for f in recursiveFileList('internal') if f.endswith('.go')]
    flist += [f for f in recursiveFileList('pkg') if f.endswith('.go')]
    flist.sort()
    return flist

def analyseFile(fname, DBType):
    cont = ""
    arrFields = []
    idxs = []
    constraints = []
    alterconstraints = ""
    f = open(fname, "r")
    reader = False
    for pos, line in enumerate(f):
        if reader == True and "}" in line:
            reader = False
            cont += generatePart(arrFields, idxs, constraints, DBType)
            arrFields = []
            idxs = []
            constraints = []
        if reader:
            arr = line.split()
            arr, idx, constraint = analyseLine(arr, structName, DBType)
            arrFields.append(arr)
            if idx: idxs.append(idx)
            if constraint: constraints.append(constraint)
            if constraint: alterconstraints += "ALTER TABLE " + structName + " ADD " + constraint + ";\n"
        if "sql generate" in line:
            structName = prevLine.split()[1].lower()+"s"
            if structName.endswith("ys"):
                structName = structName[:-2]+"ies"
            arrFields.append(structName)
            reader = True
        prevLine = line
    if DBType != 'SQLITE':
        return cont, alterconstraints
    return cont, ""

def analyseLine(x, structName, DBType):
    s = ["", ""]
    idx = False
    s[0] = x[0]
    varcharlen = False
    constraint = False
    if len(x) >= 3:
        m = re.search("varchar\((\d+|max)\)", x[2], flags=re.IGNORECASE)
        if m:
            varcharlen = m.group(1)
        if "IDX" in x[2]:
            idx = f'CREATE INDEX idx_{structName}_{x[0]} ON {structName} ({x[0]})'
    if DBType == 'SQLITE' or DBType == 'POSTGRESQL':
        s[1] = "TEXT"
        if x[1] == "int" or x[1] == "bool" or "Date" in x[1]: s[1] = "INTEGER"
        if x[1] == "float64": s[1] = "FLOAT" 
        if x[1] == "string": s[1] = "TEXT" 
        if x[1] == "[]string": s[1] = "TEXT"
        if x[1][0] == "*": s[1] = "INTEGER"
        if len(x) >= 3 and 'bigint' in x[2] and DBType == 'POSTGRESQL': s[1] = "BIGINT"
    if DBType == 'MSSQL':
        s[1] = "VARCHAR(4000)"
        if x[1] == "int" or x[1] == "bool" or "Date" in x[1]: s[1] = "INTEGER" 
        if x[1] == "float64": s[1] = "FLOAT" 
        if x[1] == "string": s[1] = "VARCHAR(4000)" 
        if x[1] == "[]string": s[1] = "VARCHAR(max)"
        if x[1][0] == "*": s[1] = "INTEGER"
        if varcharlen: s[1] = f'VARCHAR({varcharlen})'
        if len(x) >= 3 and 'bigint' in x[2]: s[1] = "BIGINT"
    if DBType == 'MYSQL':
        s[1] = "VARCHAR(4000)"
        if x[1] == "int" or x[1] == "bool" or "Date" in x[1]: s[1] = "INTEGER"
        if x[1] == "float64": s[1] = "FLOAT" 
        if x[1] == "string": s[1] = "VARCHAR(4000)" 
        if x[1] == "[]string": s[1] = "TEXT"
        if x[1][0] == "*": s[1] = "INTEGER"
        if varcharlen: s[1] = f'VARCHAR({varcharlen})'
        if varcharlen == 'max': s[1] = 'TEXT'
        if len(x) >= 3 and 'bigint' in x[2]: s[1] = "BIGINT"
    if DBType == 'ORACLE':
        s[1] = "VARCHAR(4000)"
        if x[1] == "int" or x[1] == "bool" or "Date" in x[1]: s[1] = "INTEGER"
        if x[1] == "float64": s[1] = "FLOAT" 
        if x[1] == "string": s[1] = "VARCHAR2(4000)" 
        if x[1] == "[]string": s[1] = "CLOB"
        if x[1][0] == "*":
            s[1] = "INTEGER"
        if varcharlen:
            s[1] = f'VARCHAR2({varcharlen})'
        if varcharlen == 'max':
            s[1] = 'CLOB'
    if "*" in x[1]:
        fk_table = x[1].replace("*", "").lower() + "s"
        if fk_table.endswith("ys"):
            fk_table = fk_table[:-2]+"ies"
        if "." in fk_table:
            fk_table = fk_table.split(".")[1]
    if len(x) > 2 and (x[1] == "int" or x[1] == "int64"):
        fkt = re.search("fktable\((.+)\)", x[2], flags=re.IGNORECASE)
        if fkt: fk_table = fkt.group(1)
    if len(x) >= 3:
        if 'UNIQUE' in x[2]: s[1] += ' UNIQUE'
        if 'FK_NULL' in x[2]:
            constraint = f'CONSTRAINT fk_{structName}_{x[0]} FOREIGN KEY ({x[0]}) REFERENCES {fk_table}(ID) ON DELETE SET NULL'
        if 'FK_CASCADE' in x[2]:
            constraint = f'CONSTRAINT fk_{structName}_{x[0]} FOREIGN KEY ({x[0]}) REFERENCES {fk_table}(ID) ON DELETE CASCADE'
        if 'FK_NOACTION' in x[2] and DBType == 'MSSQL':
            constraint = False
    return s, idx, constraint

def generatePart(arrFields, idxs, constraints, DBType):
    primarykeystr = {
        'SQLITE':     "INTEGER PRIMARY KEY AUTOINCREMENT",
        'MSSQL':      "INTEGER IDENTITY PRIMARY KEY",
        'MYSQL':      "INTEGER PRIMARY KEY AUTO_INCREMENT",
        'ORACLE':     "INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY",
        'POSTGRESQL': "SERIAL PRIMARY KEY",
    }
    ifnotexists = {
        'SQLITE':     "IF NOT EXISTS",
        'MSSQL':      "",
        'MYSQL':      "IF NOT EXISTS",
        'ORACLE':     "",
        'POSTGRESQL': "IF NOT EXISTS",
    }
    text = ""
    for v in arrFields:
        if type(v) is str:
            text += f'CREATE TABLE {ifnotexists[DBType]} {v}\n('
        else:
            if v[0] != "ID":
                text += f'{v[0]} {v[1]}'
            if v[0] == "ID":
                text += "ID " + primarykeystr[DBType]
            if v != arrFields[-1]:
                text += ",\n"
    if DBType == 'SQLITE':
        for constraint in constraints:
            text += ",\n"+constraint
    text += ");\n\n"
    if idxs:
        for idx in idxs:
            text += idx +";\n\n"
    return text

def createFileContent(DBType):
    cont = ""
    flist = makeFileList()
    addconstraints = ""
    for file in flist:
        contpart, alterconstraints = analyseFile(file, DBType)
        cont += contpart
        addconstraints += alterconstraints
    if addconstraints:
        cont += addconstraints
    return cont

def genFile(fname, cont):
    f = open(fname, "w") #overwrites
    f.write(cont)
    f.close()

if __name__ == "__main__":
    os.chdir(os.path.join('..', '..'))
    genFile(os.path.join("sqlscripts", "sqlite-create.sql"), createFileContent('SQLITE'))
    genFile(os.path.join("sqlscripts", "mssql-create.sql"), createFileContent('MSSQL'))
    genFile(os.path.join("sqlscripts", "mysql-create.sql"), createFileContent('MYSQL'))
    genFile(os.path.join("sqlscripts", "oracle-create.sql"), createFileContent('ORACLE'))
    genFile(os.path.join("sqlscripts", "postgresql-create.sql"), createFileContent('POSTGRESQL'))
    print("sqlscripts files generated")
