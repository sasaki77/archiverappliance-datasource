import { DefinePlugin } from 'webpack';
import type { Configuration } from 'webpack';
import { merge } from 'webpack-merge';
import grafanaConfig, { Env } from './.config/webpack/webpack.config';
import { getPackageJson } from './.config/webpack/utils';

const config = async (env: Env): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env);

  return merge(baseConfig, {
    plugins: [
      new DefinePlugin({
        __VERSION__: JSON.stringify(getPackageJson().version),
      }),
    ],
  });
};

export default config;
