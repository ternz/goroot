#!/usr/bin/env python2

import os
import sys
import glob
import shutil
import argparse
import subprocess
import platform
import time
import string
import signal

cwd = os.path.dirname(os.path.abspath(__file__))
gopath = os.getenv("GOPATH")
mobilegameserver = os.path.join(gopath, "src/mobilegameserver")
tools = os.path.join(mobilegameserver, "tools")
clientcommon = os.path.join(mobilegameserver, "clientcommon")
platcommon = os.path.join(mobilegameserver, "platcommon")
servercommon = os.path.join(mobilegameserver, "servercommon")
libgame = os.path.join(mobilegameserver, "libgame")

#need prebuild path
preBuildPath = [cwd,clientcommon,libgame]

def init():
    print("cd %s" % tools)
    os.chdir(tools)
    assert(os.system("python build.py") == 0)

    print("cd %s" % clientcommon)
    os.chdir(clientcommon)
    assert(os.system("python build.py") == 0)

    print("cd %s" % platcommon)
    os.chdir(platcommon)
    assert(os.system("python build.py") == 0)

    print("cd %s" % servercommon)
    os.chdir(servercommon)
    assert(os.system("python build.py") == 0)

    os.chdir(cwd)
    for fn in glob.glob("*.config.example"):
        fnn = fn.rpartition(".")[0]
        if not os.path.exists(fnn):
            print("cp %s -> %s" % (fn, fnn))
            shutil.copyfile(fn, fnn)

def update():
    cmd = "svn up {0}".format(os.path.split(os.getcwd())[0])
    assert(os.system(cmd) == 0)


def build():
    server = os.path.join(cwd, "gameserver")
    print("cd %s" % server)
    os.chdir(server)
    if os.system("go build -v") != 0:
        sys.exit(1)

    os.chdir(cwd)
    assert(os.system("go run export.go gameserver") == 0)



def afterbuild():
    for path in preBuildPath:
        for d,fd,fl in os.walk(path):
            for f in fl:
                sufix = os.path.splitext(f)[1][1:]
                if sufix == "go":
                     os.remove(d + '/' + f)
        for d,fd,fl in os.walk(path):
            for f in fl:
                pre,sufix = os.path.splitext(f)
                sufix = sufix[1:]
                if sufix == "goo":
                     os.rename(d + '/' + f,d+'/'+pre+".go")

def prebuild ():
    for path in preBuildPath:
        for d,fd,fl in os.walk(path):
            for f in fl:
                sufix = os.path.splitext(f)[1][1:]
                if sufix == "go":
                     dealfile(d + '/' + f,f)

def dealfile(filepath,filename):
    print "deal file:"+filepath
    strgrep1 = "Error(\""
    strgrep2 = "Debug(\""
    strgrep3 = "Info(\""
    strgrep4 = "Warning(\""
    filenameTmp = filepath+"o"
    os.rename(filepath,filenameTmp)
    input   = open(filenameTmp)
    lines   = input.readlines()
    input.close()

    output  = open(filepath,'w')
    lineNum = 0
    for line in lines:
        if not line:
            break
        lineNum += 1
        if strgrep1 in line  :
            temp    = line.split(strgrep1)
            temp1   = temp[0] +strgrep1+"["+ filename+":"+str(lineNum)+"]," + temp[1]
            #print temp
            #print temp1
            output.write(temp1)
        elif strgrep2 in line  :
            temp    = line.split(strgrep2)
            temp1   = temp[0] +strgrep2+"["+ filename+":"+str(lineNum)+"]," + temp[1]
            output.write(temp1)
        elif strgrep3 in line  :
            temp    = line.split(strgrep3)
            temp1   = temp[0] +strgrep3+"["+ filename+":"+str(lineNum)+"]," + temp[1]
            output.write(temp1)
        elif strgrep4 in line  :
            temp    = line.split(strgrep4)
            temp1   = temp[0] +strgrep4+"["+ filename+":"+str(lineNum)+"]," + temp[1]
            output.write(temp1)
        else:
            output.write(line)
    output.close()

def replace(opts):
    d, a, b = opts.replace
    cmd = 'find %s -name "*.go" | xargs grep %s | cut -d: -f1 | xargs sed -i "s|%s|%s|g"' % (d, a, a, b)
    print(cmd)
    os.system(cmd)

def version():
    release = os.path.join(cwd, "release")
    if not os.path.exists(release):
        print("mkdirs %s" % release)
        os.makedirs(release)
    cmd = "svn info | grep Revision > .svninfo"
    assert(os.system(cmd) == 0)
    with open(".svninfo") as f:
        for line in f:
            vt = "{0}-{1}-{2}".format("gameserver", line.strip('\n').split(':')[1].strip(' '), time.strftime("%Y%m%d%H%M%S", time.localtime()))
    versiondir = os.path.join(release, vt)
    if not os.path.exists(versiondir):
        print("mkdirs %s" % versiondir)
        os.makedirs(versiondir)

    server = "gameserver"
    serverdir = os.path.join(versiondir, server)
    if not os.path.exists(serverdir):
        print("mkdirs %s" % serverdir)
        os.makedirs(serverdir)

    cmd = "cp -f ./{0}/{0} {1}".format(server, serverdir)
    print(cmd)
    assert(os.system(cmd) == 0)
    cmd = "svn export {0}/clientcommon/data {1}/data".format(os.path.split(os.getcwd())[0], versiondir)
    print(cmd)
    assert(os.system(cmd) == 0)

    cmd = "cp -f ./*.*.example {0}".format(versiondir)
    print(cmd)
    assert(os.system(cmd) == 0)
    print("cd %s" % release)
    os.chdir(release)
    cmd = "tar cvjf {0}.tar.bz2 {1}".format(vt, os.path.basename(versiondir))
    assert(os.system(cmd) == 0)
    cmd = "md5sum {0}.tar.bz2 > Readme".format(vt)
    assert(os.system(cmd) == 0)
    shutil.rmtree(versiondir)

prog = os.path.basename(os.path.abspath(__file__))
parser = argparse.ArgumentParser(prog="./%s" % prog)
parser.add_argument("-i", "--init", action="store_true", dest="init", help="init build environment")
parser.add_argument("-r", "--replace", dest="replace", nargs=3, metavar="", help="execute replace for go source code")
parser.add_argument("-u", "--update", action="store_true", dest="update", help="update dependent repos")
parser.add_argument("-v", "--version", action="store_true", dest="version", help="make release version")
parser.add_argument("-l", "--log", action="store_true", dest="log", help="build with better log info")
opts = parser.parse_args()

if opts.init:
    init()
    update()
    build()
elif opts.replace:
    replace(opts)
elif opts.update:
    update()
    build()
elif opts.version:
    update()
    version()
elif opts.log:
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    signal.signal(signal.SIGINT, signal.SIG_IGN)
    prebuild()
    try:
            build()
    finally:
            afterbuild()
else:
    build()
