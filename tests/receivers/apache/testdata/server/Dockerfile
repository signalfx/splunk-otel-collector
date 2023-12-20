FROM httpd:2.4.58

RUN echo '\
ExtendedStatus on\n\
<Location /mod_status>\n\
   SetHandler server-status\n\
 </Location>\n'\
>> /usr/local/apache2/conf/httpd.conf
