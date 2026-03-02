export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType): void;
    registerRightHandSidebarComponent(
        component: React.ElementType,
        title: string,
    ): { id: string; showRHSPlugin: any };
    registerChannelHeaderButtonAction(
        icon: React.ReactElement,
        action: (channel: any) => void,
        dropdownText: string,
        tooltipText: string,
    ): void;
    registerReducer(reducer: any): void;
    registerWebSocketEventHandler(
        event: string,
        handler: (msg: any) => void,
    ): void;
}
