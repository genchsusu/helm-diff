#!/bin/sh -e

if [ -n "${HELM_LINTER_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-diff v${version} ..."

url=""
if [ "$(uname)" = "Darwin" ]; then
    url="https://github.com/genchsusu/helm-diff/releases/download/v${version}/helm-diff_${version}_darwin_amd64.tar.gz"
elif [ "$(uname)" = "Linux" ] ; then
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
        url="https://github.com/genchsusu/helm-diff/releases/download/v${version}/helm-diff_${version}_linux_arm64.tar.gz"
    else
        url="https://github.com/genchsusu/helm-diff/releases/download/v${version}/helm-diff_${version}_linux_amd64.tar.gz"
    fi
else
    url="https://github.com/genchsusu/helm-diff/releases/download/v${version}/helm-diff_${version}_windows_amd64.tar.gz"
fi

echo "$url"

mkdir -p "bin"
mkdir -p "releases/v${version}"

# Download with curl if possible.
# shellcheck disable=SC2230
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
else
    wget -q "${url}" -O "releases/v${version}.tar.gz"
fi
tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/diff" "bin/diff" || \
    mv "releases/v${version}/diff.exe" "bin/diff"
mv "releases/v${version}/plugin.yaml" .
