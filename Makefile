LIB_DIR=${INSTALL_PREFIX}/lib/systemd/system
BIN_DIR=${INSTALL_PREFIX}/usr/lib/docker/
PLUGIN_NAME=docker-volume-mem

all:
	go build -o ${PLUGIN_NAME} .


install:
	install -d -m 0755 ${LIB_DIR}
	install -m 644 systemd/${PLUGIN_NAME}.service ${LIB_DIR}
	install -d -m 0755 ${LIB_DIR}
	install -m 644 systemd/${PLUGIN_NAME}.socket ${LIB_DIR}
	install -d -m 0755 ${BIN_DIR}
	install -m 755 ${PLUGIN_NAME} ${BIN_DIR}

clean:
	rm -f ${PLUGIN_NAME}
