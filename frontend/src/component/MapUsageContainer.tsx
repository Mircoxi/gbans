import { apiGetMapUsage } from '../api';
import React, { useEffect, useMemo, useState } from 'react';
import { PieChart } from '@mui/x-charts';
import { compare, Order, RowsPerPage, stableSort } from './DataTable';
import { ContainerWithHeader } from './ContainerWithHeader';
import MapIcon from '@mui/icons-material/Map';
import Grid from '@mui/material/Unstable_Grid2';
import Box from '@mui/material/Box';
import { LoadingSpinner } from './LoadingSpinner';
import Stack from '@mui/material/Stack';
import Pagination from '@mui/material/Pagination';
import { LazyTable } from './LazyTable';
import TabContext from '@mui/lab/TabContext';
import TabList from '@mui/lab/TabList';
import Tab from '@mui/material/Tab';
import TabPanel from '@mui/lab/TabPanel';

interface MapUseChartProps {
    details: SeriesData[];
}

const MapUseChart = ({ details }: MapUseChartProps) => {
    return (
        <PieChart
            height={600}
            width={600}
            legend={{ hidden: true }}
            series={[
                {
                    data: details,
                    highlightScope: { faded: 'global', highlighted: 'item' },
                    faded: { innerRadius: 30, additionalRadius: -30 },
                    valueFormatter: (value) => {
                        return `${value.value.toFixed(2)}%`;
                    }
                }
            ]}
        />
    );
};

interface SeriesData {
    id: string;
    label: string;
    value: number;
}

interface BarChartWithTableProps {
    loading: boolean;
    data: SeriesData[];
}

const BarChartWithTable = ({ loading, data }: BarChartWithTableProps) => {
    const [sortOrder, setSortOrder] = useState<Order>('desc');
    const [sortColumn, setSortColumn] = useState<keyof SeriesData>('value');
    const [page, setPage] = useState(1);

    const rows = useMemo(() => {
        return stableSort(data, compare(sortOrder, sortColumn)).slice(
            (page - 1) * RowsPerPage.TwentyFive,
            (page - 1) * RowsPerPage.TwentyFive + RowsPerPage.TwentyFive
        );
    }, [data, page, sortColumn, sortOrder]);

    return (
        <Grid container>
            <Grid md={6} xs={12}>
                <Box
                    paddingLeft={10}
                    display="flex"
                    justifyContent="center"
                    alignItems="center"
                >
                    {loading ? (
                        <LoadingSpinner />
                    ) : (
                        <MapUseChart details={data} />
                    )}
                </Box>
            </Grid>
            <Grid md={6} xs={12}>
                {loading ? (
                    <LoadingSpinner />
                ) : (
                    <Stack>
                        <Stack direction={'row-reverse'}>
                            <Pagination
                                page={page}
                                count={Math.ceil(data.length / 25)}
                                showFirstButton
                                showLastButton
                                onChange={(_, newPage) => {
                                    setPage(newPage);
                                }}
                            />
                        </Stack>
                        <LazyTable<SeriesData>
                            columns={[
                                {
                                    label: 'Map',
                                    sortable: true,
                                    sortKey: 'label',
                                    tooltip: 'Map'
                                },
                                {
                                    label: 'Percent',
                                    sortable: true,
                                    sortKey: 'value',
                                    tooltip: 'Percentage of overall playtime',
                                    renderer: (obj) => {
                                        return obj.value.toFixed(2) + ' %';
                                    }
                                }
                            ]}
                            sortOrder={sortOrder}
                            sortColumn={sortColumn}
                            onSortColumnChanged={async (column) => {
                                setSortColumn(column);
                            }}
                            onSortOrderChanged={async (direction) => {
                                setSortOrder(direction);
                            }}
                            rows={rows}
                        />
                    </Stack>
                )}
            </Grid>
        </Grid>
    );
};

export const MapUsageContainer = () => {
    const [series, setSeries] = useState<SeriesData[]>([]);
    const [seriesMode, setSeriesMode] = useState<SeriesData[]>([]);
    const [loading, setLoading] = useState(true);

    const [value, setValue] = React.useState('1');

    const handleChange = (_: React.SyntheticEvent, newValue: string) => {
        setValue(newValue);
    };

    useEffect(() => {
        apiGetMapUsage()
            .then((resp) => {
                if (resp.result) {
                    setSeries(
                        resp.result.map((value1): SeriesData => {
                            return {
                                id: value1.map,
                                value: value1.percent,
                                label: value1.map
                            };
                        })
                    );
                    const maps: Record<string, number> = {};

                    // eslint-disable-next-line no-loops/no-loops
                    for (let i = 0; i < resp.result.length; i++) {
                        const key = resp.result[i].map
                            .replace('workshop/', '')
                            .split('_')[0];
                        if (!maps[key]) {
                            maps[key] = 0;
                        }
                        maps[key] += resp.result[i].percent;
                    }
                    const values: SeriesData[] = [];
                    // eslint-disable-next-line no-loops/no-loops
                    for (const mapsKey in maps) {
                        values.push({
                            label: mapsKey,
                            id: mapsKey,
                            value: maps[mapsKey]
                        });
                    }
                    setSeriesMode(values);
                }
            })
            .finally(() => {
                setLoading(false);
            });
    }, []);

    return (
        <ContainerWithHeader
            title={'Map Playtime Distribution'}
            iconLeft={<MapIcon />}
        >
            <TabContext value={value}>
                <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                    <TabList
                        onChange={handleChange}
                        aria-label="lab API tabs example"
                    >
                        <Tab label="By Map" value="1" />
                        <Tab label="By Mode" value="2" />
                    </TabList>
                </Box>
                <TabPanel value="1">
                    <BarChartWithTable loading={loading} data={series} />
                </TabPanel>
                <TabPanel value="2">
                    <BarChartWithTable loading={loading} data={seriesMode} />
                </TabPanel>
            </TabContext>
        </ContainerWithHeader>
    );
};
