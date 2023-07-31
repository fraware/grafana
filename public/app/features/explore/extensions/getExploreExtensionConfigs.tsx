import React from 'react';

import { PluginExtensionPoints, type PluginExtensionLinkConfig } from '@grafana/data';
import { contextSrv } from 'app/core/core';
import { dispatch } from 'app/store/store';
import { AccessControlAction } from 'app/types';

import { createExtensionLinkConfig, logWarning } from '../../plugins/extensions/utils';
import { changeCorrelationsEditorMode } from '../state/main';
import { runQueries } from '../state/query';

import { AddToDashboardForm } from './AddToDashboard/AddToDashboardForm';
import { getAddToDashboardTitle } from './AddToDashboard/getAddToDashboardTitle';
import { type PluginExtensionExploreContext } from './ToolbarExtensionPoint';

export function getExploreExtensionConfigs(): PluginExtensionLinkConfig[] {
  try {
    return [
      createExtensionLinkConfig<PluginExtensionExploreContext>({
        title: 'Add to dashboard',
        description: 'Use the query and panel from explore and create/add it to a dashboard',
        extensionPointId: PluginExtensionPoints.ExploreToolbarAction,
        icon: 'apps',
        category: 'Dashboards',
        configure: () => {
          const canAddPanelToDashboard =
            contextSrv.hasAccess(AccessControlAction.DashboardsCreate, contextSrv.isEditor) ||
            contextSrv.hasAccess(AccessControlAction.DashboardsWrite, contextSrv.isEditor);

          // hide option if user has insufficient permissions
          if (!canAddPanelToDashboard) {
            return undefined;
          }

          return {};
        },
        onClick: (_, { context, openModal }) => {
          openModal({
            title: getAddToDashboardTitle(),
            body: ({ onDismiss }) => <AddToDashboardForm onClose={onDismiss!} exploreId={context?.exploreId!} />,
          });
        },
      }),
      createExtensionLinkConfig<PluginExtensionExploreContext>({
        title: 'Add Correlation',
        description: 'Create a correlation from this query',
        extensionPointId: PluginExtensionPoints.ExploreToolbarAction,
        icon: 'link',
        configure: () => {
          return {};
        },
        onClick: (_, { context }) => {
          dispatch(changeCorrelationsEditorMode({ correlationsEditorMode: true }));
          if (context?.exploreId) {
            dispatch(runQueries({ exploreId: context?.exploreId }));
          }
        },
      }),
    ];
  } catch (error) {
    logWarning(`Could not configure extensions for Explore due to: "${error}"`);
    return [];
  }
}