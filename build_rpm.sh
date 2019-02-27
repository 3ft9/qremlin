#!/bin/bash
set -ex

NAME=qremlin
TOPDIR=/tmp/${NAME}_rpm_build_topdir
rm -rf ${TOPDIR}
mkdir -p ${TOPDIR}/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
chcon -Rt svirt_sandbox_file_t ${TOPDIR}

## TODO: Add a version to the code and grab it here!
VERSION="1" # $(sed -n "s/^.*__version__ = '\(.*\)'.*/\1/p" zib/version.py)

mkdir -p /tmp/qremlin-${VERSION}
cp -r * /tmp/qremlin-${VERSION}

pushd /tmp && tar zcvf ${TOPDIR}/SOURCES/${NAME}-${VERSION}.tar.gz ${NAME}-${VERSION} && popd
rm -rf ${NAME}-${VERSION}

cp ${NAME}.rpmspec ${TOPDIR}/SPECS
sed -i "s/{{VERSION}}/$VERSION/" ${TOPDIR}/SPECS/${NAME}.rpmspec

rpmbuild --define "_topdir $TOPDIR" -ba ${TOPDIR}/SPECS/${NAME}.rpmspec

mv ${TOPDIR}/RPMS/x86_64/${NAME}-${VERSION}-1.x86_64.rpm .

rm -rf ${TOPDIR}
