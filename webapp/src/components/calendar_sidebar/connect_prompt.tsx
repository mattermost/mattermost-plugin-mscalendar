import {useCallback} from 'react';

import {MattermostTheme} from '@/utils/calendar_theme';

import CalendarIconSVG from './calendar_icon_svg';

const CONNECT_USING_BROWSER_MESSAGE = 'Please connect your Microsoft Calendar account using your web browser. The desktop app cannot open the OAuth window directly.';

function isDesktopApp(): boolean {
    const userAgent = window.navigator.userAgent;
    return userAgent.indexOf('Mattermost') !== -1 && userAgent.indexOf('Electron') !== -1;
}

interface ConnectPromptProps {
    theme: MattermostTheme;
    pluginServerRoute: string;
    sendEphemeralPost: (message: string) => void;
}

const ConnectPrompt = ({theme, pluginServerRoute, sendEphemeralPost}: ConnectPromptProps) => {
    const handleConnect = useCallback((e: React.MouseEvent<HTMLAnchorElement>) => {
        e.preventDefault();
        if (isDesktopApp()) {
            sendEphemeralPost(CONNECT_USING_BROWSER_MESSAGE);
            return;
        }
        window.open(
            `${pluginServerRoute}/oauth2/connect`,
            'Connect Mattermost to Microsoft Calendar',
            'height=570,width=520,noopener,noreferrer',
        );
    }, [pluginServerRoute, sendEphemeralPost]);

    return (
        <div className='mscalendar-sidebar__connect'>
            <div className='mscalendar-sidebar__connect-welcome'>
                {'Welcome to Microsoft Calendar'}
            </div>
            <div className='mscalendar-sidebar__connect-icon'>
                <CalendarIconSVG theme={theme}/>
            </div>
            <div className='mscalendar-sidebar__connect-prompt'>
                {'Connect your account'}
                <br/>
                {'to get started'}
            </div>
            <a
                className='mscalendar-sidebar__connect-btn'
                href={`${pluginServerRoute}/oauth2/connect`}
                onClick={handleConnect}
                style={{
                    backgroundColor: theme.buttonBg,
                    color: theme.buttonColor,
                }}
            >
                {'Connect account'}
            </a>
        </div>
    );
};

export default ConnectPrompt;
