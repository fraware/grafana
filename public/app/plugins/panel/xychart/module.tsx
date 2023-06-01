import { PanelPlugin } from '@grafana/data';
import { defaultFieldConfig } from '@grafana/schema/src/raw/composable/xychart/panelcfg/x/XYChartPanelCfg_types.gen';
import { commonOptionsBuilder } from '@grafana/ui';

import { AutoEditor } from './AutoEditor';
import { ManualEditor } from './ManualEditor';
import { XYChartPanel2 } from './XYChartPanel2';
import { getScatterFieldConfig } from './config';
import { Options, ScatterFieldConfig } from './types';

export const plugin = new PanelPlugin<Options, ScatterFieldConfig>(XYChartPanel2)
  .useFieldConfig(getScatterFieldConfig(defaultFieldConfig))
  .setPanelOptions((builder) => {
    builder
      .addRadio({
        path: 'seriesMapping',
        name: 'Series mapping',
        defaultValue: 'auto',
        settings: {
          options: [
            { value: 'auto', label: 'Auto', description: 'No changes to saved model since 8.0' },
            { value: 'manual', label: 'Manual' },
          ],
        },
      })
      .addCustomEditor({
        id: 'xyPlotConfig',
        path: 'dims',
        name: '',
        editor: AutoEditor,
        showIf: (cfg) => cfg.seriesMapping === 'auto',
      })
      .addCustomEditor({
        id: 'series',
        path: 'series',
        name: '',
        defaultValue: [],
        editor: ManualEditor,
        showIf: (cfg) => cfg.seriesMapping === 'manual',
      });

    commonOptionsBuilder.addTooltipOptions(builder);
    commonOptionsBuilder.addLegendOptions(builder);
  });
