#!/usr/bin/env sh
# ゲートアプリ(Android)をDockerでビルドする。ホストにJDKやAndroid SDKは不要。
#
#   ./build.sh                                # APKのビルドと単体テスト
#   ./build.sh assembleDebug                  # 任意のGradleタスク
#
# 出力: app/build/outputs/apk/debug/app-debug.apk
#
# Gradleのキャッシュとデバッグ用の署名鍵は、Dockerボリューム gate-gradle-cache に保存する。
# 署名鍵が毎回変わると端末で上書きインストールできなくなるため、このボリュームは消さないこと。
set -eu
cd "$(dirname "$0")"

docker volume inspect gate-gradle-cache >/dev/null 2>&1 || {
  docker volume create gate-gradle-cache >/dev/null
  docker run --rm -v gate-gradle-cache:/cache alpine chown "$(id -u):$(id -g)" /cache
}

if [ $# -eq 0 ]; then
  set -- assembleDebug testDebugUnitTest
fi

exec docker run --rm -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GRADLE_USER_HOME=/cache -e ANDROID_USER_HOME=/cache/android \
  -v gate-gradle-cache:/cache -v "$PWD":/work -w /work \
  ghcr.io/cirruslabs/android-sdk:35 ./gradlew --no-daemon "$@"
