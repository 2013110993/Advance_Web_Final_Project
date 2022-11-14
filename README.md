# Advance_Web_Final_Project


#Add User
export BODY='{"name":"federico rosado", "email":"federico@gmail.com", "password":"letmeintest"}'
curl -d "$BODY" localhost:4000/v1/users

#Activate User
export BODY='{"token": "CIWMQAPE4HVFPJCFPKWMMGDRY4"}'
curl -X PUT -d "$BODY" localhost:4000/v1/user/activate
