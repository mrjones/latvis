include $(GOROOT)/src/Make.inc

all: clean install

TARG=latvis

DIRS=\
				location\
				latitude\
				visualization\
				server\
				localserver\
        cristimap\

TEST=\
				location\
				server\
				visualization\


clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))
test.dirs: $(addsuffix .test, $(TEST))

%.clean:
				+cd $* && gomake clean

%.install:
				+cd $* && gomake install

%.nuke:
				+cd $* && gomake nuke

%.test:
				+cd $* && gomake test

clean: clean.dirs

install: install.dirs

test:   test.dirs

nuke: nuke.dirs
				rm -rf "$(GOROOT)"/pkg/*

appengine_local: install
				dev_appserver.py --address mrjon.es .

appengine_deploy: install
				appcfg.py update .
