import React, {
    ChangeEvent,
    SyntheticEvent,
    useCallback,
    useState
} from 'react';
import { ProfileSelectionInput } from './ProfileSelectionInput';
import Stack from '@mui/material/Stack';
import {
    apiCreateBanCIDR,
    BanPayloadCIDR,
    BanReason,
    BanReasons,
    banReasonsList,
    BanType,
    Duration,
    Durations,
    IAPIBanCIDRRecord,
    ip2int,
    SteamID
} from '../api';
import { ConfirmationModal, ConfirmationModalProps } from './ConfirmationModal';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormHelperText from '@mui/material/FormHelperText';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { useUserFlashCtx } from '../contexts/UserFlashCtx';
import IPCIDR from 'ip-cidr';
import { Heading } from './Heading';

export interface BanCIDRModalProps
    extends ConfirmationModalProps<IAPIBanCIDRRecord> {
    reportId?: number;
    targetId?: SteamID;
}

export const BanCIDRModal = ({
    open,
    setOpen,
    onSuccess,
    targetId
}: BanCIDRModalProps) => {
    const [targetSteamId, setTargetSteamId] = useState<SteamID>(
        targetId ?? BigInt(0)
    );
    const [input, setInput] = useState<SteamID>(BigInt(0));
    const [duration, setDuration] = useState<Duration>(Duration.dur48h);
    const [customDuration, setCustomDuration] = useState<string>('');
    const [banReason, setBanReason] = useState<BanReason>(BanReason.Cheating);
    const [noteText, setNoteText] = useState<string>('');
    const [reasonText, setReasonText] = useState<string>('');
    const [network, setNetwork] = useState<string>('');
    const [networkSize, setNetworkSize] = useState<number>(0);
    const { sendFlash } = useUserFlashCtx();

    const handleSubmit = useCallback(() => {
        if (!targetSteamId) {
            sendFlash('error', 'no steamId');
            return;
        }
        if (banReason == BanReason.Custom && customDuration == '') {
            sendFlash('error', 'Custom duration cannot be empty');
            return;
        }
        const dur = duration == Duration.durCustom ? customDuration : duration;
        if (!dur) {
            sendFlash('error', 'Custom duration invalid');
            return;
        }
        const opts: BanPayloadCIDR = {
            target_id: targetSteamId,
            duration: dur,
            network: network,
            reason_text: reasonText,
            reason: banReason,
            note: noteText,
            ban_type: BanType.Banned
        };

        apiCreateBanCIDR(opts)
            .then((ban) => {
                sendFlash(
                    'success',
                    `CIDR ban created successfully: ${ban.net_id}`
                );
                onSuccess && onSuccess(ban);
            })
            .catch((err) => {
                sendFlash('error', `Failed to create ban: ${err}`);
            });
    }, [
        targetSteamId,
        banReason,
        customDuration,
        duration,
        network,
        reasonText,
        noteText,
        sendFlash,
        onSuccess
    ]);

    const handleUpdateNetwork = (evt: SyntheticEvent) => {
        const value = (evt.target as HTMLInputElement).value;
        setNetwork(value);
        if (value !== '') {
            try {
                const cidr = new IPCIDR(value);
                if (cidr != undefined) {
                    setNetworkSize(
                        ip2int(cidr?.end()) - ip2int(cidr?.start()) + 1
                    );
                }
            } catch (e) {
                if (e instanceof TypeError) {
                    // TypeError on invalid input we can ignore
                } else {
                    throw e;
                }
            }
        }
    };

    const handleUpdateReason = (evt: SelectChangeEvent<BanReason>) => {
        setBanReason(evt.target.value as BanReason);
    };

    const handleUpdateDuration = (evt: SelectChangeEvent<Duration>) => {
        setDuration(evt.target.value as Duration);
    };

    const handleUpdateNote = (evt: ChangeEvent<HTMLInputElement>) => {
        setNoteText((evt.target as HTMLInputElement).value);
    };

    return (
        <ConfirmationModal
            open={open}
            setOpen={setOpen}
            onSuccess={() => {
                setOpen(false);
            }}
            onCancel={() => {
                setOpen(false);
            }}
            onAccept={() => {
                handleSubmit();
            }}
            aria-labelledby="modal-title"
            aria-describedby="modal-description"
        >
            <Stack spacing={2}>
                <Heading>Ban CIDR Range</Heading>
                {!targetId && (
                    <ProfileSelectionInput
                        fullWidth
                        onProfileSuccess={(profile) => {
                            if (profile) {
                                setTargetSteamId(profile.player.steam_id);
                            } else {
                                setTargetSteamId(BigInt(0));
                            }
                        }}
                        input={input}
                        setInput={setInput}
                    />
                )}
                <Stack spacing={3} alignItems={'center'}>
                    <TextField
                        fullWidth={true}
                        id={'network'}
                        label={'Network Range (CIDR Format)'}
                        onChange={handleUpdateNetwork}
                    />
                    <Typography variant={'body1'}>
                        Current number of hosts in range: {networkSize}
                    </Typography>

                    <Select<BanReason>
                        fullWidth
                        labelId="reason-label"
                        id="reason-helper"
                        value={banReason}
                        onChange={handleUpdateReason}
                    >
                        {banReasonsList.map((v) => (
                            <MenuItem key={`time-${v}`} value={v}>
                                {BanReasons[v]}
                            </MenuItem>
                        ))}
                    </Select>
                    {banReason == BanReason.Custom && (
                        <FormControl fullWidth>
                            <InputLabel id="reasonText-label">
                                Reason
                            </InputLabel>
                            <TextField
                                id={'reasonText'}
                                value={reasonText}
                                onChange={(evt) => {
                                    setReasonText(evt.target.value);
                                }}
                            />
                        </FormControl>
                    )}
                    <FormControl fullWidth>
                        <InputLabel id="duration-label">
                            Ban Duration
                        </InputLabel>
                        <Select<Duration>
                            fullWidth
                            labelId="duration-label"
                            id="duration-helper"
                            value={duration}
                            onChange={handleUpdateDuration}
                        >
                            {Durations.map((v) => (
                                <MenuItem key={`time-${v}`} value={v}>
                                    {v}
                                </MenuItem>
                            ))}
                        </Select>
                        <FormHelperText>
                            Choosing custom will allow you to input a custom
                            duration
                        </FormHelperText>
                    </FormControl>

                    {duration == Duration.durCustom && (
                        <FormControl fullWidth>
                            <InputLabel id="customDuration-label">
                                Custom Duration
                            </InputLabel>
                            <TextField
                                id={'customDuration'}
                                value={customDuration}
                                onChange={(evt) => {
                                    setCustomDuration(evt.target.value);
                                }}
                            />
                        </FormControl>
                    )}

                    <FormControl fullWidth>
                        <TextField
                            id="note-field"
                            label="Moderator Notes (hidden from public)"
                            multiline
                            value={noteText}
                            onChange={handleUpdateNote}
                            rows={10}
                            variant="outlined"
                        />
                    </FormControl>
                </Stack>
            </Stack>
        </ConfirmationModal>
    );
};
