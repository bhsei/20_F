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
        try:
            self.db_connection = pymysql.connect(
                    host = host,
                    port = port,
                    user = user,
                    password = password,
                    database = database,
                    charset = 'utf8mb4'
                    )
        except pymysql.Error as e:
            print(e)
            self.db_connection = None
            return

        db = self.db_connection

        cursor = db.cursor()
        cursor.execute("CREATE TABLE IF NOT EXISTS USER("
                       "ID INT," 
                       "createTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
                       ")")
        cursor.close()

    def db_insert_user_record(self, dict):
        """Insert New Record into the USER

        :param ID: identifier, Primary Key of Identifier
        :param dict: The key-value collection intends to insert into USER
        :return:
        """
        if self.db_connection is None:
            return -1

        db = self.db_connection
        cursor = db.cursor()

        ls = [(k, v) for k, v in dict.items() if v is not None]
        stmt = 'INSERT USER (' + ','.join([i[0] for i in ls]) + ') ' \
               'VALUES (' + ','.join(repr(i[1]) for i in ls) + ');'
        # print(stmt)

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError:
            db.rollback()
            return -3

        return 1

    def db_update_user_record(self, ID, dict):
        """Update Record For User Table

        :param ID: Index, Primary Key for the table USER
        :param dict: The key-value dictionary intends to insert into USER as field-value pairs
        :return:
             1 Sucessfully
            -1 DataBase Connected Failed
            -2 Checking Table USER Failed
            -3 SQL Statement Error or No these Fields
        """

        if self.db_connection is None:
            return -1

        db = self.db_connection
        cursor = db.cursor()

        ls = [(k, v) for k, v in dict.items() if v is not None]
        tmp = ""
        for i in ls:
            tmp += i[0] + "=" + i[1] + ","
        tmp = tmp[:-1]
        stmt = 'UPDATE USER SET ' + tmp + ' WHERE id={}'.format(ID)

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError:
            db.rollback()
            return -3

        return 1

    def db_add_setting(self, dict):
        """Add Fields For User Table

        :param dict: The key-value dictionary intends to insert into USER as setting-type pairs
        :return:
             1 Successfully
            -1 DataBase Connected Failed
            -2 Checking Table USER Failed
            -3 SQL Statement Error or No these Fields
        """
        if self.db_connection is None:
            return -1

        db = self.db_connection
        cursor = db.cursor()

        exist_settings = []
        repeat_settings = {}
        new_settings = {}

        if self.db_table_if_exist(cursor, "USER"):
            try:
                cursor.execute("SELECT * FROM USER")
                for field in cursor.description:
                    exist_settings.append(field[0])
            except pymysql.Error as e:
                print(e.args[0], e.args[1])
                return -2

        tmp = ""
        for k, v in dict.items():
            if k in exist_settings:
                repeat_settings[k] = v
            else:
                new_settings[k] = v
                tmp += "ADD "+ k + " " + v + ","

        stmt = "ALTER TABLE USER " + tmp[:-1]
        # print(stmt)

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError:
            db.rollback()
            return -3

        return 1
        # return {
        #     "repeat_setting" : repeat_settings,
        #     "new_setting": new_settings
        # }

    def db_del_setting(self, dict):
        """

        :param dict: the setting, shown as field-type pair, intends to be deleted
        :return:
             1 Successfully
            -1 DataBase Connected Failed
            -2 Checking Table USER Failed
            -3 SQL Statement Error or No these Fields
        """
        if self.db_connection is None:
            return -1

        db = self.db_connection
        cursor = db.cursor()

        exist_settings = []
        non_exist_settings = {}
        del_settings = {}
        if self.db_table_if_exist(cursor, "USER"):
            try:
                cursor.execute("SELECT * FROM USER")
                for field in cursor.description:
                    exist_settings.append(field[0])
            except pymysql.Error as e:
                print(e.args[0], e.args[1])
                return -2

        # print(exist_settings)
        tmp = ""
        for k, v in dict.items():
            # print(k, k in exist_settings)
            if k not in exist_settings:
                non_exist_settings[k] = v
            else:
                del_settings[k] = v
                tmp += "DROP " + k + ","
        stmt = "ALTER TABLE USER " + tmp[:-1]
        # print(stmt)

        try:
            cursor.execute(stmt)
            db.commit()
        except pymysql.MySQLError:
            db.rollback()
            return -3

        # cursor.execute("SELECT * FROM USER")
        # print(cursor.description)
        return 1

        # return {
        #     "repeat_setting" : repeat_settings,
        #     "new_setting": new_settings
        # }

    def de_reset(self):
        """ Reset the User Table for Test

        :return:
        """
        db = self.db_connection
        cursor = db.cursor()
        try:
            cursor.execute("DROP TABLE USER")
            db.commit()
        except pymysql.Error:
            db.rollback()
        finally:
            cursor.close()

    def db_check(self):
        """ Check the User Table

        :return: the ResultSet
        """
        db = self.db_connection
        cursor = db.cursor()
        try:
            cursor.execute("SELECT * FROM USER")
            return cursor.fetchall()
        except pymysql.Error:
            db.rollback()
            return None
        finally:
            cursor.close()

    def db_table_if_exist(self, cursor, table):
        """ Checking Table IF Exist

        :param cursor: cursor in the connection
        :param table: table name
        :return:
            True, Table exist
            False, Table does not exist
        """
        sql = "show tables"
        cursor.execute(sql)

        tables = cursor.fetchall()
        tables_list = re.findall('(\'.*?\')', str(tables))
        tables_list = [re.sub("'", '', each) for each in tables_list]
        # print(tables, tables_list)

        if table in tables_list:
            # print("table exists")
            return True
        else:
            # print("table not exists")
            return False
            # try:
            #     sql = createTable[0]
            #     # print(sql)
            #     cursor.execute(sql)
            #     db.commit()
            #     print("Successfully added table")
            # except:
            #     db.rollback()
            #     print("UnSuccessfully added table")

    def db_query(self, id, cols = None):
        """ the Query api

        :param id: the given Row to query
        :param cols: the given fields to return
        :return:
            None empty dict or invalid cols collection
            dict the query fieldName-fieldValue as key-value

        """
        if self.db_connection is None:
            raise pymysql.Error

        db = self.db_connection
        cursor = db.cursor()

        exist_fields = []
        res = {}
        try:
            cursor.execute("SELECT * FROM USER WHERE ID = {}".format(id))
            query_all = cursor.fetchall()
            if len(query_all) == 0:
                return None
            for field in cursor.description:
                exist_fields.append(field[0])
        except pymysql.Error as e:
            raise e

        if cols is None:
            cols = exist_fields

        for i in range(len(exist_fields)):
            if exist_fields[i] in cols:
                res[exist_fields[i]] = query_all[0][i]

        return res

if __name__ == "__main__":
    config = {
        "host": "127.0.0.1",
        "port": 3306,
        "user": "root",
        "password": "123456",
        "database": "test",
    }

    demo = DBOperation(config)
    if demo.db_connection is None:
        print("No~~~~~~~~~~~~~~~~~~")
    else:
        print("Success!!")

    # demo.db_alter_user_setting(1, 1)
    print("Insert-----", demo.db_insert_user_record({
        "ID": 2,
    }))

    print("Update----", demo.db_update_user_record(2, {
        "new4": "100"
    }))

    print("Add Settings----", demo.db_add_setting({
        "new4": "INT",
        "new5": "CHAR(10)"
    }))

    print("Del Settings----", demo.db_del_setting({
        "new5": "INT",
        "new6": "CHAR(10)"
    }))
    demo.db_check()
    # demo.de_reset()
    demo.db_connection.close()
