all: 
	@./build.sh
clean:
	@rm -f cloth-physics
install: all
	@cp cloth-physics /usr/local/bin
uninstall: 
	@rm -f /usr/local/bin/cloth-physics
package:
	@NOCOPY=1 ./build.sh package