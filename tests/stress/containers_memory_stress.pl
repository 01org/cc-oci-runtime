#!/usr/bin/perl -w

#  This file is part of cc-oci-runtime.
#  #
#  #  Copyright (C) 2016 Intel Corporation
#  #
#  #  This program is free software; you can redistribute it and/or
#  #  modify it under the terms of the GNU General Public License
#  #  as published by the Free Software Foundation; either version 2
#  #  of the License, or (at your option) any later version.
#  #
#  #  This program is distributed in the hope that it will be useful,
#  #  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  #  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  #  GNU General Public License for more details.
#  #
#  #  You should have received a copy of the GNU General Public License
#  #  along with this program; if not, write to the Free Software
#  #  Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
#
#  #  Description of the test:
#  #  This test launches 10 containers and put them in a start and stop loop of thousand of iterations  
 

my $REPS=5000;
my $SRATE=5;
my $MAX_CONTAINERS=10;
my $workdir = `pwd -P`;
my $t1=0;
my $t2=0;
my $delta_Ms=0;
my $DATE = qx(date +%m/%d/%Y);
my $TIME = qx(date +%H:%M:%S);
chomp($DATE);
chomp($TIME);
chomp($workdir);
$workdir .='/data/';
my $COLLECT_LOG="mem_$MAX_CONTAINERS"."_containers_$REPS"."_reps_srate_$SRATE";
my $TIMELINE_LOG="mem_$MAX_CONTAINERS"."_containers_$REPS"."_reps_srate_$SRATE"."_$DATE"."_$TIME.log";


sub setup {
    print "Creating data directories...\n";
    for (1..$MAX_CONTAINERS) {
    	`mkdir -p $workdir/s$_`;
    }
    system "echo \"Iteration,Date,Time,ElapsedT(ms)\" > ${TIMELINE_LOG}";
    print "Launching collectl...\n";
    system "collectl -oTm -sZm -i1:${SRATE} --procfilt cdocker,cqemu,ccc-proxy,ccc-shim --procopts m -f ${COLLECT_LOG} &";
}    
sub docker_ps {
    system('docker ps -a');
}
sub docker_stop {
    my $r = `docker stop \`docker ps -aq\``;
    $r = `docker rm -f \`docker ps -aq\``;
    print $r;
}
sub docker_run {
    my ($iter) = @_;
    print "Running ...$iter\n";    
    for (1..$MAX_CONTAINERS) {
	`docker run -v $workdir/s$_:/tdir -tid clearlinux bash -c 'while true; do echo $iter > /tdir/out.log; sleep 1; done'`
    }
}
sub write_delta {
    $DATE = qx(date +%m/%d/%Y);
    $TIME = qx(date +%H:%M:%S);
    chomp($DATE);
    chomp($TIME);
    $t2 = qx(date +%s%N);
    $delta_ms = ($t2 - $t1) / 1000000;
    system "echo \"$_,${DATE},${TIME},${delta}\" >> ${TIMELINE_LOG}";
}
sub cleanup {
  system('ps axf | grep collectl | grep -v grep | awk \'{ print "kill -TERM -- - " $1}\' | sh');
}


setup();
`sleep 10`;
for (1..$REPS) {
    print "\tRunning $_ iteration\n";
    $t1 = qx(date +%s%N);
    &docker_run($_);
    &docker_ps();
    `sleep 1`;
    &docker_stop();
}
`sleep 5`;
&cleanup();    
