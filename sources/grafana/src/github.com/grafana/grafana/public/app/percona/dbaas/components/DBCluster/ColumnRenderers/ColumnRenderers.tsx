import React from 'react';
import { DATABASE_LABELS } from 'app/percona/shared/core';
import { DBClusterConnection } from '../DBClusterConnection/DBClusterConnection';
import { DBClusterStatus } from '../DBClusterStatus/DBClusterStatus';
import { DBCluster, DBClusterStatus as Status } from '../DBCluster.types';
import { DBClusterParameters } from '../DBClusterParameters/DBClusterParameters';
import { DBClusterName } from '../DBClusterName/DBClusterName';
import { DBClusterActions } from '../DBClusterActions/DBClusterActions';
import { DBClusterActionsProps } from '../DBClusterActions/DBClusterActions.types';

export const clusterNameRender = (dbCluster: DBCluster) => <DBClusterName dbCluster={dbCluster} />;

export const databaseTypeRender = (dbCluster: DBCluster) => DATABASE_LABELS[dbCluster.databaseType];

export const clusterStatusRender = (dbCluster: DBCluster) => {
  const { status, message, finishedSteps, totalSteps } = dbCluster;

  return (
    <DBClusterStatus
      status={status || Status.changing}
      message={message}
      finishedSteps={finishedSteps}
      totalSteps={totalSteps}
    />
  );
};

export const connectionRender = (dbCluster: DBCluster) => <DBClusterConnection dbCluster={dbCluster} />;
export const parametersRender = (dbCluster: DBCluster) => <DBClusterParameters dbCluster={dbCluster} />;

export const clusterActionsRender = ({
  setSelectedCluster,
  setDeleteModalVisible,
  setEditModalVisible,
  getDBClusters,
}: Omit<DBClusterActionsProps, 'dbCluster'>) => (dbCluster: DBCluster) => (
  <DBClusterActions
    dbCluster={dbCluster}
    setSelectedCluster={setSelectedCluster}
    setDeleteModalVisible={setDeleteModalVisible}
    setEditModalVisible={setEditModalVisible}
    getDBClusters={getDBClusters}
  />
);
