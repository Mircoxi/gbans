import React, { useCallback, useEffect, useState } from 'react';
import Grid from '@mui/material/Unstable_Grid2';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';
import MenuItem from '@mui/material/MenuItem';
import {
    apiGetMessages,
    apiGetServers,
    MessageQuery,
    PersonMessage,
    Server
} from '../api';
import { steamIdQueryValue } from '../util/text';
import Button from '@mui/material/Button';
import ButtonGroup from '@mui/material/ButtonGroup';
import { Heading } from '../component/Heading';
import { LazyTable } from '../component/LazyTable';
import { logErr } from '../util/errors';
import { Order, RowsPerPage } from '../component/DataTable';
import { formatISO9075 } from 'date-fns/fp';
import { TablePagination } from '@mui/material';
import { useTimer } from 'react-timer-hook';
import ChatIcon from '@mui/icons-material/Chat';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Box from '@mui/material/Box';
import { DelayedTextInput } from '../component/DelayedTextInput';

const anyServer: Server = {
    server_name: 'Any',
    server_id: 0,
    server_name_long: 'Any',
    address: '',
    port: 27015,
    longitude: 0.0,
    latitude: 0.0,
    is_enabled: true,
    cc: '',
    default_map: '',
    password: '',
    rcon: '',
    players_max: 24,
    region: '',
    reserved_slots: 8,
    updated_on: new Date(),
    created_on: new Date()
};

interface ChatQueryState<T> {
    startDate: Date | null;
    endDate: Date | null;
    steamId: string;
    nameQuery: string;
    messageQuery: string;
    sortOrder: Order;
    sortColumn: keyof T;
    page: number;
    rowPerPageCount: number;
    selectedServer: number;
}

const localStorageKey = 'chat_query_state';

const loadState = () => {
    let config: ChatQueryState<PersonMessage> = {
        startDate: null,
        endDate: null,
        sortOrder: 'desc',
        sortColumn: 'person_message_id',
        selectedServer: anyServer.server_id,
        rowPerPageCount: RowsPerPage.Fifty,
        nameQuery: '',
        messageQuery: '',
        steamId: '',
        page: 0
    };
    const item = localStorage.getItem(localStorageKey);
    if (item) {
        config = JSON.parse(item);
    }
    return config;
};

export const AdminChat = () => {
    const init = loadState();
    const [startDate, setStartDate] = useState<Date | null>(init.startDate);
    const [endDate, setEndDate] = useState<Date | null>(init.endDate);
    const [steamId, setSteamId] = useState<string>(init.steamId);
    const [nameQuery, setNameQuery] = useState<string>(init.nameQuery);
    const [messageQuery, setMessageQuery] = useState<string>(init.messageQuery);
    const [sortOrder, setSortOrder] = useState<Order>(init.sortOrder);
    const [sortColumn, setSortColumn] = useState<keyof PersonMessage>(
        init.sortColumn
    );
    const [servers, setServers] = useState<Server[]>([]);
    const [rows, setRows] = useState<PersonMessage[]>([]);
    const [page, setPage] = useState(init.page);
    const [rowPerPageCount, setRowPerPageCount] = useState<number>(
        init.rowPerPageCount
    );
    const [refreshTime, setRefreshTime] = useState<number>(0);
    const [totalRows, setTotalRows] = useState<number>(0);
    //const [pageCount, setPageCount] = useState<number>(0);

    const [nameValue, setNameValue] = useState<string>(init.nameQuery);
    const [steamIDValue, setSteamIDValue] = useState<string>(init.steamId);
    const [messageValue, setMessageValue] = useState<string>(init.messageQuery);

    const [selectedServer, setSelectedServer] = useState<number>(
        init.selectedServer
    );

    const curTime = new Date();
    curTime.setSeconds(curTime.getSeconds() + refreshTime);

    const { isRunning, restart } = useTimer({
        expiryTimestamp: curTime,
        autoStart: false
    });

    const saveState = useCallback(() => {
        localStorage.setItem(
            localStorageKey,
            JSON.stringify({
                endDate,
                steamId,
                messageQuery,
                nameQuery,
                page,
                rowPerPageCount,
                selectedServer,
                sortColumn,
                sortOrder,
                startDate
            } as ChatQueryState<PersonMessage>)
        );
    }, [
        endDate,
        messageQuery,
        nameQuery,
        page,
        rowPerPageCount,
        selectedServer,
        sortColumn,
        sortOrder,
        startDate,
        steamId
    ]);

    useEffect(() => {
        apiGetServers().then((resp) => {
            if (!resp.status || !resp.result) {
                return;
            }
            setServers([
                anyServer,
                ...resp.result.sort((a, b) => {
                    return a.server_name.localeCompare(b.server_name);
                })
            ]);
        });
    }, []);

    const restartTimer = useCallback(() => {
        if (refreshTime <= 0) {
            return;
        }
        const newTime = new Date();
        newTime.setSeconds(newTime.getSeconds() + refreshTime);
        restart(newTime, true);
    }, [refreshTime, restart]);

    useEffect(() => {
        if (isRunning) {
            // wait for timer to exec
            return;
        }
        const opts: MessageQuery = {};
        if (selectedServer > 0) {
            opts.server_id = selectedServer;
        }
        if (nameQuery) {
            opts.persona_name = nameQuery;
        }
        if (messageQuery) {
            opts.query = messageQuery;
        }
        if (steamId) {
            opts.steam_id = steamId;
        }
        if (startDate) {
            opts.sent_after = startDate;
        }
        if (endDate) {
            opts.sent_before = endDate;
        }
        opts.limit = rowPerPageCount;
        opts.offset = page * rowPerPageCount;
        opts.order_by = sortColumn;
        opts.desc = sortOrder == 'desc';
        apiGetMessages(opts)
            .then((resp) => {
                const count = resp.result?.total_messages || 0;
                setRows(resp.result?.messages || []);
                setTotalRows(count);
            })
            .catch((e) => {
                logErr(e);
            });
        saveState();
        restartTimer();
    }, [
        endDate,
        messageQuery,
        nameQuery,
        page,
        rowPerPageCount,
        selectedServer,
        sortColumn,
        sortOrder,
        startDate,
        steamId,
        isRunning,
        restart,
        restartTimer,
        saveState
    ]);

    const reset = () => {
        setNameQuery('');
        setNameValue('');
        setSteamId('');
        setSteamIDValue('');
        setSelectedServer(anyServer.server_id);
        setStartDate(null);
        setEndDate(null);
        setPage(0);
        setRefreshTime(0);
    };

    return (
        <Grid container spacing={2} paddingTop={3}>
            <Grid xs={12}>
                <Paper elevation={1}>
                    <Stack>
                        <Heading iconLeft={<FilterAltIcon />}>
                            Chat Filters
                        </Heading>

                        <Grid
                            container
                            padding={2}
                            spacing={2}
                            justifyContent={'center'}
                            alignItems={'center'}
                        >
                            <Grid xs={6} md={3}>
                                <DelayedTextInput
                                    value={nameValue}
                                    setValue={setNameValue}
                                    placeholder={'Name'}
                                    onChange={(value) => {
                                        setNameQuery(value);
                                    }}
                                />
                            </Grid>
                            <Grid xs={6} md={3}>
                                <DelayedTextInput
                                    value={steamIDValue}
                                    setValue={setSteamIDValue}
                                    placeholder={'Steam ID'}
                                    onChange={(value) => {
                                        setSteamId(value);
                                    }}
                                />
                            </Grid>
                            <Grid xs={6} md={3}>
                                <DelayedTextInput
                                    value={messageValue}
                                    setValue={setMessageValue}
                                    placeholder={'Message'}
                                    onChange={(value) => {
                                        setMessageQuery(value);
                                    }}
                                />
                            </Grid>
                            <Grid xs={6} md={3}>
                                <Select<number>
                                    fullWidth
                                    value={selectedServer}
                                    onChange={(event) => {
                                        servers
                                            .filter(
                                                (s) =>
                                                    s.server_id ==
                                                    event.target.value
                                            )
                                            .map((s) =>
                                                setSelectedServer(s.server_id)
                                            );
                                    }}
                                    label={'Server'}
                                >
                                    {servers.map((server) => {
                                        return (
                                            <MenuItem
                                                value={server.server_id}
                                                key={server.server_id}
                                            >
                                                {server.server_name}
                                            </MenuItem>
                                        );
                                    })}
                                </Select>
                            </Grid>

                            <Grid xs={6} md={3}>
                                <DesktopDatePicker
                                    sx={{ width: '100%' }}
                                    label="Date Start"
                                    format={'MM/dd/yyyy'}
                                    value={startDate}
                                    onChange={(newValue: Date | null) => {
                                        setStartDate(newValue);
                                    }}
                                />
                            </Grid>
                            <Grid xs={6} md={3}>
                                <DesktopDatePicker
                                    sx={{ width: '100%' }}
                                    label="Date End"
                                    format="MM/dd/yyyy"
                                    value={endDate}
                                    onChange={(newValue: Date | null) => {
                                        setEndDate(newValue);
                                    }}
                                />
                            </Grid>
                            <Grid xs md={3} mdOffset="auto">
                                <ButtonGroup size={'large'} fullWidth>
                                    <Button
                                        variant={'contained'}
                                        color={'info'}
                                        onClick={reset}
                                    >
                                        Reset
                                    </Button>
                                </ButtonGroup>
                            </Grid>
                        </Grid>
                    </Stack>
                </Paper>
            </Grid>

            <Grid
                xs={12}
                container
                justifyContent="space-between"
                alignItems="center"
                flexDirection={{ xs: 'column', sm: 'row' }}
            >
                <Grid xs={3}>
                    <Box sx={{ width: 120 }}>
                        <FormControl fullWidth>
                            <InputLabel id="auto-refresh-label">
                                Auto-Refresh
                            </InputLabel>
                            <Select<number>
                                labelId="auto-refresh-label"
                                id="auto-refresh"
                                label="Auto Refresh"
                                value={refreshTime}
                                onChange={(
                                    event: SelectChangeEvent<number>
                                ) => {
                                    setRefreshTime(
                                        event.target.value as number
                                    );

                                    restartTimer();
                                }}
                            >
                                <MenuItem value={0}>Off</MenuItem>
                                <MenuItem value={10}>5s</MenuItem>
                                <MenuItem value={15}>15s</MenuItem>
                                <MenuItem value={30}>30s</MenuItem>
                                <MenuItem value={60}>60s</MenuItem>
                            </Select>
                        </FormControl>
                    </Box>
                </Grid>
                <Grid xs={'auto'}>
                    <TablePagination
                        component="div"
                        variant={'head'}
                        page={page}
                        count={totalRows}
                        showFirstButton
                        showLastButton
                        rowsPerPage={rowPerPageCount}
                        onRowsPerPageChange={(
                            event: React.ChangeEvent<
                                HTMLInputElement | HTMLTextAreaElement
                            >
                        ) => {
                            setRowPerPageCount(
                                parseInt(event.target.value, 10)
                            );
                            setPage(0);
                        }}
                        onPageChange={(_, newPage) => {
                            setPage(newPage);
                        }}
                    />
                </Grid>
            </Grid>

            <Grid xs={12}>
                <Heading iconLeft={<ChatIcon />}>Chat Messages</Heading>
                <LazyTable<PersonMessage>
                    sortOrder={sortOrder}
                    sortColumn={sortColumn}
                    onSortColumnChanged={async (column) => {
                        setSortColumn(column);
                    }}
                    onSortOrderChanged={async (direction) => {
                        setSortOrder(direction);
                    }}
                    columns={[
                        {
                            label: 'Server',
                            tooltip: 'Server',
                            sortKey: 'server_id',
                            align: 'left',
                            width: 100,
                            onClick: (o) => {
                                setSelectedServer(o.server_id);
                            },
                            queryValue: (o) =>
                                `${o.server_id} + ${o.server_name}`,
                            renderer: (row) => (
                                <Typography variant={'button'}>
                                    {row.server_name}
                                </Typography>
                            )
                        },
                        {
                            label: 'Created',
                            tooltip: 'Time the message was sent',
                            sortKey: 'created_on',
                            sortType: 'date',
                            align: 'left',
                            width: 180,
                            queryValue: (o) => steamIdQueryValue(o.steam_id),
                            renderer: (row) => (
                                <Typography variant={'body1'}>
                                    {`${formatISO9075(row.created_on)}`}
                                </Typography>
                            )
                        },
                        {
                            label: 'Name',
                            tooltip: 'Persona Name',
                            sortKey: 'persona_name',
                            width: 250,
                            align: 'left',
                            onClick: (o) => {
                                setSteamId(o.steam_id);
                                setSteamIDValue(o.steam_id);
                            },
                            queryValue: (o) => `${o.persona_name}`,
                            renderer: (row) => (
                                <Typography variant={'body2'}>
                                    {row.persona_name}
                                </Typography>
                            )
                        },
                        {
                            label: 'Message',
                            tooltip: 'Message',
                            sortKey: 'body',
                            align: 'left',
                            queryValue: (o) => o.body
                        }
                    ]}
                    rows={rows}
                />
            </Grid>
        </Grid>
    );
};
