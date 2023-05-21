#!/usr/bin/env bash

set -e

NAME="zabbix"
VERSION="${1}"
WDIR=$(pwd)
if [ "x${VERSION}" == "x" ]; then
  echo "Please specify version as first argument"
  exit 1
fi

case ${OSTYPE} in
    darwin*)
        _SHA256='shasum -a 256 '

        ;;
    linux-gnu*)
        _SHA256='sha256sum'
        ;;
    *)
        echo "Unsupported 'OSTYPE'='${OSTYPE}'"
        exit 1
        ;;
esac

_build()
{

  export OS="${1}"
  export ARCH="${2}"

  export GOOS="${OS}"
  export GOARCH="${ARCH}"

  echo "Building package for ${OS} ${ARCH}"

  make build
  if [ -f ${GOPATH}/bin/terraform-provider-zabbix ]; then
    cd ${GOPATH}/bin
  else
    cd ${GOPATH}/bin/${OS}_${ARCH}
  fi

  mv terraform-provider-${NAME} terraform-provider-${NAME}_v${VERSION}
  zip terraform-provider-${NAME}_${VERSION}_${OS}_${ARCH}.zip terraform-provider-${NAME}_v${VERSION}
  mv terraform-provider-${NAME}_${VERSION}_${OS}_${ARCH}.zip  ${WDIR}/release/${VERSION}/
  cd -
}

mkdir -p ${WDIR}/release/${VERSION}
cat > ${WDIR}/release/${VERSION}/terraform-provider-${NAME}_${VERSION}_manifest.json <<EOF
{
  "version": 1,
  "metadata": {
    "protocol_versions": ["5.0"]
  }
}
EOF

_build linux 386
_build linux amd64
_build linux arm64
_build darwin amd64
_build darwin arm64

cd ${WDIR}/release/${VERSION}/
${_SHA256} *.zip > terraform-provider-${NAME}_${VERSION}_SHA256SUMS
gpg --detach-sign terraform-provider-${NAME}_${VERSION}_SHA256SUMS
