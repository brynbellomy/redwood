import React, { useState, useCallback, useRef } from 'react'
import styled from 'styled-components'
import Modal, { ModalTitle, ModalContent, ModalActions } from '../Modal'
import Input from '../Input'
import Button from '../Button'
import CurrentUserAvatar from '../CurrentUserAvatar'
import useRedwood from '../../hooks/useRedwood'
import useModal from '../../hooks/useModal'
import useAPI from '../../hooks/useAPI'
import useNavigation from '../../hooks/useNavigation'
import useStateTree from '../../hooks/useStateTree'

const SUserControlContainer = styled.div`
    display: flex;
    align-items: center;
    height: 56px;
    width: 100%;
    background-color: ${props => props.theme.color.grey[500]};
`

const SUserLeft = styled.div`
    width: calc(250px - 12px);
    display: flex;
    align-items: center;
    padding-left: 12px;
    transition: .15s ease-in-out all;
    height: 100%;
    cursor: pointer;

    &:hover {
        background: ${props => props.theme.color.grey[300]};
    }
`

const UsernameWrapper = styled.div`
    width: 100px;
    display: flex;
    flex-direction: column;
    margin-left: 8px;
`

const Username = styled.div`
    overflow-x: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: white;
    font-weight: 700;
    font-size: 0.8rem;
`

const NodeAddress = styled.div`
    overflow-x: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 10px;
    color: rgba(255, 255, 255, .6);
    font-weight: 300;
`

const SCurrentUserAvatar = styled(CurrentUserAvatar)`
    height: 40px;
`

function UserControl() {
    let { onPresent, onDismiss } = useModal('user profile')
    let { nodeAddress } = useRedwood()
    let { selectedServer } = useNavigation()
    let registry = useStateTree(`${selectedServer}/registry`)
    nodeAddress = !!nodeAddress ? nodeAddress.toLowerCase() : null
    let userPhotoURL
    let username
    if (registry && registry.users && registry.users[nodeAddress]) {
        username = registry.users[nodeAddress].username
        if (registry.users[nodeAddress].photo) {
            userPhotoURL = `http://localhost:8080/users/${nodeAddress}/photo?state_uri=${selectedServer}/registry`
        }
    }
    return (
        <SUserControlContainer>
            <SUserLeft onClick={onPresent}>
                <SCurrentUserAvatar />
                <UsernameWrapper>
                    <Username>{!!username ? username : nodeAddress}</Username>
                    <NodeAddress>{!!username ? nodeAddress : null}</NodeAddress>
                </UsernameWrapper>
            </SUserLeft>
            <UserProfileModal onDismiss={onDismiss} />
        </SUserControlContainer>
    )
}

function UserProfileModal({ onDismiss }) {
    const [username, setUsername] = useState('')
    const { nodeAddress } = useRedwood()
    const api = useAPI()
    const { selectedServer } = useNavigation()
    const photoFileRef = useRef()

    const onSave = useCallback(async () => {
        if (!api) { return }
        try {
            let photoFile
            if (photoFileRef && photoFileRef.current && photoFileRef.current.files && photoFileRef.current.files.length > 0) {
                photoFile = photoFileRef.current.files[0]
            }
            await api.updateProfile(nodeAddress, selectedServer, username, photoFile)
            onDismiss()
        } catch (err) {
            console.error(err)
        }
    }, [api, nodeAddress, selectedServer, username, onDismiss])

    const onChangeUsername = useCallback((e) => {
        if (e.code === 'Enter') {
            onSave()
        } else {
            setUsername(e.target.value)
        }
    }, [onSave, setUsername])

    return (
        <Modal modalKey="user profile">
            <ModalTitle>Your profile</ModalTitle>
            <ModalContent>
                <div>
                    <input type="file" ref={photoFileRef} />
                </div>
                <div>
                    Username:
                    <Input value={username} onChange={onChangeUsername} />
                </div>
            </ModalContent>
            <ModalActions>
                <Button onClick={onSave} primary>Save</Button>
                <Button onClick={onDismiss}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

export default UserControl