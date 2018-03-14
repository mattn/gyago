Gyago
=====

Gyazo application written in go

Client
------

Currently, This is broken.
For working correctly, you should patch CL 4635063.
http://codereview.appspot.com/4635063/ . gyago is image uploader. it upload image file specified as argument to gyazo.com.


Server
------

http://go-gyazo.appspot.com/
You can upload image to go-gyazo server with following command.

    gyago -e=http://go-gyago.appspot.com imagefile.png


