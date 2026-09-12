/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

export const CARD_HOVER_SX = {
  boxShadow: 'none',
  transition: 'all 0.2s ease-in-out',
  cursor: 'pointer',
  '&:hover': {
    boxShadow: (theme: { palette: { primary: { main: string } } }) => `0px 0px 0px 1.5px ${theme.palette.primary.main}`,
  },
} as const;

export const PROVIDER_ICON_SX = {
  width: 40,
  height: 40,
  color: 'text.secondary',
  '&:hover': { color: 'primary.main', bgcolor: 'action.hover' },
} as const;

/**
 * GitHub's brand guidelines require its mark to stay black/white — never
 * recolored on hover — so this omits PROVIDER_ICON_SX's hover color swap.
 */
export const GITHUB_ICON_SX = {
  width: 40,
  height: 40,
  color: 'text.primary',
  '&:hover': { bgcolor: 'action.hover' },
} as const;

/** Colors a required TextField's asterisk red. */
export const REQUIRED_FIELD_SX = { '& .MuiFormLabel-asterisk': { color: 'error.main' } } as const;

/** Label sitting above a card, outside it — the option's name in the create flows. */
export const SECTION_LABEL_SX = {
  color: 'text.secondary',
  fontWeight: 500,
} as const;
