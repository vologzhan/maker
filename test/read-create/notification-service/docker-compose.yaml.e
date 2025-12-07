services:
  notification-db:
    container_name: notification-db
    image: postgres:18
    ports:
      - ${POSTGRES_PORT}:5432
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
    volumes:
      - db_data:/var/lib/postgresql

volumes:
  db_data:
