FROM scratch

COPY etc etc
COPY themes themes
COPY custom-error-pages /

ENV ERROR_FILES_PATH /themes/knockout

CMD ["/custom-error-pages"]
