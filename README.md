# Team Honey Badger Slide Thingy-mabob
Alright, there are some things you are going to need for this to work...
1. A Computer (duh) (actually will need a computer with WIFI) (RPI 3B+ if using raspberry pi)
2. A TV/Monitor (duh)
3. A few more than 2 Brain Cells (may be difficult to obtain)
4. A secure internet connection to compensate for the fact that I did not bother to implement TLS (duh)

If you plan on using a raspberry pi you will need to follow [these](https://www.raspberrypi.com/documentation/computers/getting-started.html) instructions to set the computer up, otherwise, figure it out on your own.

Once your computer is set-up, now we need to install the dependencies for this application. If you are running a linux based OS use this command, otherwise, lock-in I guess.
```bash
apt-get install libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev
```
I think there is a dependency missing from this list, so if you get linker errors, just put it into ChatGPT and it'll probably figure it out idk.

Anyways, now we need the [Go Compiler](https://go.dev/doc/install) and [Git](https://git-scm.com/install/linux).

Next navigate to your documents folder throught the terminal or through the file manager (once you get there right-click and open in terminal(if thats not an option, look it up idk)).

Once in the documents folder, run the following command...
```bash
git clone https://github.com/zb1ndev/thbc.git
```

Now navigate to the root of the project you just cloned, and run the following commands...
```bash
cd ./src
go mod tidy
cd ..
chmod +x ./build.sh
./build.sh
```

If it doesnt run at this point, I probably missed a dependency, as ChatGPT about it, it'll probably work eventually.