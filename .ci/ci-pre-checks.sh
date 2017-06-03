#!/bin/bash
#  This file is part of cc-oci-runtime.
#
#  Copyright (C) 2017 Intel Corporation
#
#  This program is free software; you can redistribute it and/or
#  modify it under the terms of the GNU General Public License
#  as published by the Free Software Foundation; either version 2
#  of the License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program; if not, write to the Free Software
#  Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

#---------------------------------------------------------------------
# Description: Perform basic checks on the branch before attempting
#   to run the build and test suites. If this script fails, the CI run
#   should be aborted.
#---------------------------------------------------------------------

repo="github.com/clearcontainers/tests/cmd/checkcommits"
go get -u "$repo"

#---------------------------------------------------------------------
# FIXME: debug

set -x
env
pwd
ls -al
git status
git branch
git remote -v
git --no-pager log -5

git rev-list --reverse master..HEAD
git rev-list --reverse master..

git rev-list --no-merges --reverse master..HEAD
git rev-list --no-merges --reverse master..

git remote -v
git reflog

#---------------------------------------------------------------------

checkcommits --verbose --need-fixes --need-sign-offs

#---------------------------------------------------------------------
