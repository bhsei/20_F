import pymysql
import re


class DBOperation:
    """
    @desc DataBase Operation
    @date 5/8
    """

    db_connection = None

    def db_init(self, host, port, user, password, database):
        """ Another Way of Initial

        :param host: Database mapping
        :param port: the port where the Database running on
        :param user: the Database account
        :param password: the Database password
        :param database: choose
        :return:
        """
        self.db_connection = pymysql.connect(
            host=host,
            port=port,
            user=user,
            password=password,
            database=database,
            charset='utf8mb4'
        )

        db = self.db_connection

        cursor = db.cursor()
        cursor.execute("CREATE TABLE IF NOT EXISTS USER("
                       "ID INT,"
                       "createTime BIGINT"
                       ")")
        cursor.close()
    
    def _del_user(self, ID):
        db = self.db_connection
        cursor = db.cursor()
        try:
            cursor.execute("DELETE FROM USER WHERE ID=%s", ID)
            db.commit()
        except pymysql.MySQLError as e:
            db.rollback()
            raise e
        return 1
    
    def _get_user_timestamp(self, ID):
        db = self.db_connection
        cursor = db.cursor()

        exist_records = []

        cursor.execute("SELECT ID, createTime FROM USER WHERE id=%s", ID)
        exist_records = cursor.fetchone()
        
        if exist_records:
            return exist_records[1]
        return None

    def _user_exists(self, ID, timestamp):
        old = self._get_user_timestamp(ID)
        if not old:
            return False
        if old < timestamp:
            self._del_user(ID)
            return False
        return True

    def db_del_user(self, ID, timestamp):
        if self.db_connection is None:
            raise pymysql.MySQLError("mysql uninitialized")
        old_timestamp = self._get_user_timestamp(ID)
        if old_timestamp and old_timestamp <= timestamp:
            self._del_user(ID)
        return 1

    def db_insert_or_update(self, ID, settings, timestamp):
        """ Insert Or Update Record
        :param ID: identifier, Primary Key of Identifier
        :param dict: The key-value collection intends to insert into USER
        :return:
        """
        if self.db_connection is None:
            raise pymysql.MySQLError("mysql uninitialized")

        db = self.db_connection
        cursor = db.cursor()

        if self._user_exists(ID, timestamp):
            keys = settings.keys()
            settings_stm = list(map(lambda k: "{}=%s".format(k), keys))
            values = list(map(lambda k: settings[k], keys))
            values += [ID]
            settings_stm = ",".join(settings_stm)
            stmt = "UPDATE USER SET {} WHERE ID=%s".format(settings_stm)
            try:
                cursor.execute(stmt, values)
                db.commit()
                return 1
            except pymysql.MySQLError as e:
                db.rollback()
                raise e
            return -1

        keys = settings.keys()
        stm_keys = ["ID", "createTime"] + list(keys)
        stm_vals = ",".join(["%s"] * len(stm_keys))
        stm_keys = ",".join(stm_keys)
        vals = [ID, timestamp] + list(map(lambda k: settings[k], keys))
        stmt = "INSERT USER ({}) VALUES ({})".format(stm_keys, stm_vals)
        try:
            cursor.execute(stmt, vals)
            db.commit()
        except pymysql.MySQLError as e:
            db.rollback()
            raise e
        return 1

    def db_add_setting(self, settings):
        """Add Fields For User Table

        :param dict: The key-value dictionary intends to insert into USER as setting-type pairs
        :return:
             1 Successfully
            -1 DataBase Connected Failed
            -2 Checking Table USER Failed
            -3 SQL Statement Error or No these Fields
        """
        if self.db_connection is None:
            raise pymysql.MySQLError("mysql uninitialized")

        db = self.db_connection
        cursor = db.cursor()

        cursor.execute("SELECT * FROM USER")
        description = list(map(lambda k: k[0], cursor.description))
        settings = filter(lambda k: k not in description, settings)
        settings = map(lambda k: "ADD {} VARCHAR(256)".format(k), settings)
        stmt = "ALTER TABLE USER {}".format(",".join(settings))

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError as e:
            db.rollback()
            raise e

        return 1

    def db_del_setting(self, settings):
        """

        :param dict: the setting, shown as field-type pair, intends to be deleted
        :return:
             1 Successfully
            -1 DataBase Connected Failed
            -2 Checking Table USER Failed
            -3 SQL Statement Error or No these Fields
        """
        if self.db_connection is None:
            raise pymysql.MySQLError("mysql uninitialized")

        db = self.db_connection
        cursor = db.cursor()

        cursor.execute("SELECT * FROM USER")

        description = list(map(lambda k: k[0], cursor.description))
        settings = filter(lambda k: k in description, settings)
        settings = map(lambda k: "DROP {}".format(k), settings)
        stmt = "ALTER TABLE USER {}".format(",".join(settings))

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError as e:
            db.rollback()
            raise e

        return 1

    def db_query_setting(self, cols):
        if self.db_connection is None:
            raise pymysql.Error("mysql uninitialized")
        db = self.db_connection
        cursor = db.cursor()

        q = map(lambda k: "{} = '{}'".format(k, cols[k]), cols.keys())
        stmt = "SELECT ID FROM USER WHERE {}".format(" AND ".join(q))
        cursor.execute(stmt)
        query_all = cursor.fetchall()
        return list(map(lambda q: q[0], query_all))

    def db_query(self, id, timestamp, cols):
        """ the Query api

        :param id: the given Row to query
        :param cols: the given fields to return
        :return:
            None empty dict or invalid cols collection
            dict the query fieldName-fieldValue as key-value

        """
        if self.db_connection is None:
            raise pymysql.Error("mysql uninitialized")

        db = self.db_connection
        cursor = db.cursor()

        if not self._user_exists(id, timestamp):
            return {}

        res = {}
        cursor.execute("SELECT * FROM USER WHERE ID = {}".format(id))
        query_result = cursor.fetchone()
        description = list(map(lambda k: k[0], cursor.description))

        for col in cols:
            res[col] = ""

        for i in range(len(description)):
            if description[i] in cols:
                res[description[i]] = query_result[i]

        return res
