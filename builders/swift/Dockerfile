FROM swift

WORKDIR /src
ONBUILD COPY . /src
ONBUILD RUN swift build -c release

ENV PORT {{PORT}}
EXPOSE {{PORT}}

CMD ["/bin/bash", "-c", "find .build -executable -type f -not -name '*.so'| swift package describe | grep "Type: executable" | head | bash"]
