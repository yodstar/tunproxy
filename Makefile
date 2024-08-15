TARGET := tunproxy

.PHONY  : all clean clean rebuild install

all     : $(TARGET)

build   : all

rebuild : clean all

test    :

clean   :
	rm -fv build/$(TARGET)

update : 
	install -m0755 ./tunproxy /usr/local/tunproxy/tunproxy
	systemctl restart $(TARGET) && systemctl status $(TARGET)

install :
	install -d -m0755 /usr/local/tunproxy
	install -m0755 ./build/tunproxy /usr/local/tunproxy/tunproxy
	install -m0600 ./build/tunproxy.conf /usr/local/tunproxy/tunproxy.conf
	install -m0755 ./tunproxy.service /usr/lib/systemd/system/tunproxy.service
	systemctl daemon-reload

$(TARGET) :
	# go build -o build/$(TARGET) -race -ldflags "-w -s" $(TARGET)
	go build -o build/$(TARGET) -ldflags "-w -s" $(TARGET)
