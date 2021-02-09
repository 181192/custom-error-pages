FROM scratch

COPY etc etc
COPY www www
COPY custom-error-pages /

ENV ERROR_FILES_PATH /www

CMD ["/custom-error-pages"]
