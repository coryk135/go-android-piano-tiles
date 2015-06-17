docker run -v $GOPATH/src:/src golang/mobile /bin/bash -c 'cd /src/github.com/coryk135/go-android && ./make.bash'
adb uninstall com.example.basic
adb install -r bin/nativeactivity-debug.apk

adb shell am start -a android.intent.action.MAIN \
	-n com.example.basic/android.app.NativeActivity
