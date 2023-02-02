FROM amd64/alpine:3.14
ENV APP_HOME /game-app
ENV RULE_TAG develop
RUN mkdir -p $APP_HOME
WORKDIR $APP_HOME
COPY ./load-test $APP_HOME/load-test
COPY ./entrypoint.sh $APP_HOME/entrypoint.sh
RUN chmod +x $APP_HOME/load-test
RUN chmod +x $APP_HOME/entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]
