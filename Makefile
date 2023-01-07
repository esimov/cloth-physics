all: 
	@./build.sh
clean:
	@rm -f gio-cloth
install: all
	@cp gio-cloth /usr/local/bin
uninstall: 
	@rm -f /usr/local/bin/gio-cloth
package:
	@NOCOPY=1 ./build.sh package