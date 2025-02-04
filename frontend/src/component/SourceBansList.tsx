import { apiGetSourceBans, sbBanRecord } from '../api';
import React, { useEffect, useState, JSX } from 'react';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableContainer from '@mui/material/TableContainer';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import Stack from '@mui/material/Stack';

interface SourceBansListProps {
    steam_id: string;
    is_reporter: boolean;
}

export const SourceBansList = ({
    steam_id,
    is_reporter
}: SourceBansListProps): JSX.Element => {
    const [bans, setBans] = useState<sbBanRecord[]>([]);
    useEffect(() => {
        apiGetSourceBans(steam_id).then((resp) => {
            if (resp.result) {
                setBans(resp.result);
            }
        });
    }, [steam_id]);

    if (!bans.length) {
        return <></>;
    }

    return (
        <Paper elevation={1}>
            <Stack padding={2} spacing={1}>
                <Typography variant={'h5'}>
                    {is_reporter
                        ? 'Reporter SourceBans History'
                        : 'Suspect SourceBans History'}
                </Typography>
                <TableContainer>
                    <Table size="small">
                        <TableHead>
                            <TableRow>
                                <TableCell>Created</TableCell>
                                <TableCell>Source</TableCell>
                                <TableCell>Name</TableCell>
                                <TableCell>Reason</TableCell>
                                <TableCell>Permanent</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {bans.map((ban) => {
                                return (
                                    <TableRow key={`ban-${ban.ban_id}`}>
                                        <TableCell>{ban.created_on}</TableCell>
                                        <TableCell>{ban.site_name}</TableCell>
                                        <TableCell>
                                            {ban.persona_name}
                                        </TableCell>
                                        <TableCell>{ban.reason}</TableCell>
                                        <TableCell>
                                            {ban.permanent ? 'True' : 'False'}
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Stack>
        </Paper>
    );
};
