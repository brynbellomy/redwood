import React, { Fragment, useState, useContext } from 'react'
import styled from 'styled-components'
import { Link } from 'react-router-dom'

import Input, { InputLabel } from './../Input'
import Button from './../Button'

const SAccount = styled.section`
  background-color: ${props => props.theme.color.grey[500]};
  height: 100vh;
  width: 100vw;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
`

const SAccountHeader = styled.div`
  height: 250px;
  background: transparent;
  width: 100%;
`

const SAccountCard = styled.div`
  width: 450px;
  background: #36393f;
  box-shadow: 0 2px 10px 0 rgb(0 0 0 / 20%);
  border-radius: 4px;
  padding-bottom: 24px;
`

const SAccountCardHeader = styled.h2`
  color: white;
  font-size: 28px;
  margin-top: 24px;
  margin-bottom: 8px;
  text-align: center;
`

const SAccountCardDesc = styled.p`
  color: rgba(255, 255, 255, .8);
  font-size: 12px;
  text-align: center;
`

const SAccountCardContent = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  padding: 16px;
`

const SLink = styled(Link)`
  font-size: 10px;
  color: #635bff;
  margin-top: 8px;
`

function SignIn(props) {

  const onSignIn = async () => {
    console.log('Logged In')
    try {
      let resp = await (await fetch('http://localhost:54231/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          mnemonic: props.mnemonic,
        }),
      })).json()

      console.log(resp)
      props.setIsLoggedIn(true)
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <Fragment>
      <InputLabel
        label={'Mnemonic'}
      >
        <Input
          value={props.mnemonic}
          onChange={(event) => props.setMnemonic(event.currentTarget.value)}
          type={'password'}
        />
      </InputLabel>
      <SLink to={'/signup'}>Create an account.</SLink>
      <Button
        onClick={onSignIn}
        primary
        style={{ width: '100%', marginTop: 12 }}
        disabled={!props.mnemonic}
      >Sign In</Button>
    </Fragment>
  )
}

function Account(props) {
  const [accountState, setAccountState] = useState('signin')
  const [mnemonic, setMnemonic] = useState('')

  return (
    <SAccount>
      {/* <SAccountHeader /> */}
      <SAccountCard>
        <SAccountCardHeader>Sign In</SAccountCardHeader>
        <SAccountCardDesc>Always keep your mnemonic safe.</SAccountCardDesc>
        <SAccountCardContent>
          <SignIn
            mnemonic={mnemonic}
            setMnemonic={setMnemonic}
            setIsLoggedIn={props.setIsLoggedIn}
          />
        </SAccountCardContent> 
      </SAccountCard>
    </SAccount>
  )
}

export default Account