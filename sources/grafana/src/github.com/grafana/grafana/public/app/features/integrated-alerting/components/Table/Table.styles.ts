import { css } from 'emotion';
import { GrafanaTheme } from '@grafana/data';

export const getStyles = (theme: GrafanaTheme) => {
  const { colors } = theme;

  const borderColor = colors.border2;
  const backgroundColorBody = colors.bg1;
  const backgroundColorHeader = colors.bg2;

  return {
    /* This will make the table scrollable when it gets too small */
    tableWrap: css`
      border: 1px solid ${borderColor};
      display: block;
      max-width: 100%;
    `,
    table: css`
      /* This is required to make the table full-width */
      display: block;
      max-width: 100%;

      table {
        /* Make sure the inner table is always as wide as needed */
        width: 100%;
        border-spacing: 0;

        thead {
          tr {
            height: 48px;

            th {
              position: sticky;
              top: 0;
              cursor: pointer;
            }
          }
        }

        tbody {
          tr {
            height: 70px;
          }
        }

        tr {
          :last-child {
            td {
              border-bottom: 0;
            }
          }
        }
        th,
        td {
          background-color: ${backgroundColorBody};
          margin: 0;
          padding: 0 16px;
          border-bottom: 1px solid ${borderColor};
          border-right: 1px solid ${borderColor};

          :last-child {
            border-right: 0;
          }
        }

        th {
          background-color: ${backgroundColorHeader};
        }
      }

      .pagination {
        padding: 0.5rem;
      }
    `,
  };
};
