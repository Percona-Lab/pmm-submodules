import { useEffect, useState } from 'react';
import { logger } from '@percona/platform-core';
import { FulfilledPromiseResult, processPromiseResults } from 'app/percona/shared/helpers/promises';
import { Databases } from 'app/percona/shared/core';
import { Kubernetes } from '../Kubernetes/Kubernetes.types';
import { DBCluster, GetDBClustersAction, DBClusterPayload, OperatorDatabasesMap } from './DBCluster.types';
import { Operators } from './AddDBClusterModal/DBClusterBasicOptions/DBClusterBasicOptions.types';
import { KubernetesOperatorStatus } from '../Kubernetes/OperatorStatusItem/KubernetesOperatorStatus/KubernetesOperatorStatus.types';
import { newDBClusterService } from './DBCluster.utils';

const RECHECK_INTERVAL = 10000;
const DATABASES = [Databases.mysql, Databases.mongodb];

const OPERATORS: Partial<OperatorDatabasesMap> = {
  [Databases.mysql]: Operators.xtradb,
  [Databases.mongodb]: Operators.psmdb,
};

export const useDBClusters = (kubernetes: Kubernetes[]): [DBCluster[], GetDBClustersAction, boolean] => {
  const [dbClusters, setDBClusters] = useState<DBCluster[]>([]);
  const [loading, setLoading] = useState(false);
  let timer: NodeJS.Timeout;

  const getDBClusters = async (triggerLoading = true) => {
    if (triggerLoading) {
      setLoading(true);
    }

    try {
      const requests = DATABASES.map(database => getClusters(kubernetes, database));
      const results = await Promise.all(requests);
      const clustersList = results.reduce((acc, r) => acc.concat(r), []);

      setDBClusters(clustersList);
    } catch (e) {
      logger.error(e);
    } finally {
      if (triggerLoading) {
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    getDBClusters();

    timer = setInterval(() => getDBClusters(false), RECHECK_INTERVAL);

    return () => clearTimeout(timer);
  }, []);

  return [dbClusters, getDBClusters, loading];
};

const getClusters = async (kubernetes: Kubernetes[], databaseType: Databases): Promise<DBCluster[]> => {
  const dbClusterService = newDBClusterService(databaseType);
  const kubernetesByOperator = kubernetes.filter(kubernetesCluster => {
    const operator = OPERATORS[databaseType] as Operators;

    return kubernetesCluster.operators[operator].status === KubernetesOperatorStatus.ok;
  });
  const requests = kubernetesByOperator.map(dbClusterService.getDBClusters);
  const results = await processPromiseResults(requests);

  const clustersList: DBCluster[] = results.reduce((acc: DBCluster[], r, index) => {
    if (r.status !== 'fulfilled') {
      return acc;
    }

    const clusters: DBClusterPayload[] = (r as FulfilledPromiseResult).value?.clusters ?? [];

    // eslint-disable-next-line arrow-body-style
    const resultClusters = clusters.map(cluster => {
      return dbClusterService.toModel(cluster, kubernetesByOperator[index].kubernetesClusterName, databaseType);
    });

    return acc.concat(resultClusters);
  }, []);

  return clustersList;
};
