FROM nginx:alpine

# copy configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY sites-enabled /etc/nginx/sites-enabled
COPY includes /etc/nginx/includes
COPY certs /etc/certs

EXPOSE 80
