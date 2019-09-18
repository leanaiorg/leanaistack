#!/bin/bash
DOMAIN='*.dev-003.scaleout.se'
sudo docker run -it --rm --name certbot \
    -v "$(pwd)/dev-003/:/etc/letsencrypt" \
            certbot/certbot -d $DOMAIN --manual --preferred-challenges dns certonly

cp live/$DOMAIN/fullchain.pem ../
cp live/$DOMAIN/privkey.pem ../
