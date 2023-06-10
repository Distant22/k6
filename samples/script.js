import { sleep, check, group } from 'k6'
import http from 'k6/http'

export const options = {
  vus: 1000,
  duration: '10s',
}


export default function Register() {

  function generateRandomString(length) {
    var result = '';
    var characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
  }

  let response
  const data = generateRandomString(10)
  group('Register', function () {
    response = http.post( 
      `http://${__ENV.JPETSTORE_IP}:8080/jpetstore/actions/Account.action;`,
      `username=${data}&password=j2ee&repeatedPassword=j2ee&account.firstName=${data}&account.lastName=${data}&account.email=${data}&account.phone=${data}&account.address1=${data}&account.address2=${data}&account.city=${data}&account.state=${data}&account.zip=${data}&account.country=${data}&account.languagePreference=english&account.favouriteCategoryId=FISH&newAccount=Save Account Information&_sourcePage=oXxdIE9wLvknHfdpOfVv1_HSSP7U0TjQXKUueZ-pcEl12_lEHVTjmySBuR9aRaQkB9cwk2viOSxiuKfc8m-IwFgxcfO1dOlqFZCyM249tTA=&__fp=rvB7Mww1bCdoEMnIev_TleDqL3938m9SADN3W-ia4AF5Jc9HqbsskLB5Jvjsathjuu3sXzxWsGB6gPRLGCAYz7mAsO_etUGKgPBinwNCCwtQw_WxpkK9tqy7bzWeGQtdEWFx82zrF8Z19IQ2XtnRRWNNV0G-8ZTVErw2oYgh9uerKQnfYYTQx3ZRAWaUn90q`,
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      }
    )
    check(response, { 'status equals 200': response => response.status.toString() === '200' })
  })

  // Automatically added sleep
  sleep(1)
}